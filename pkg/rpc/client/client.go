package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/neophora/neo2go/pkg/core/state"
	"github.com/neophora/neo2go/pkg/core/transaction"
	"github.com/neophora/neo2go/pkg/crypto/keys"
	"github.com/neophora/neo2go/pkg/rpc/request"
	"github.com/neophora/neo2go/pkg/rpc/response"
	"github.com/neophora/neo2go/pkg/util"
	"github.com/pkg/errors"
)

const (
	defaultDialTimeout    = 4 * time.Second
	defaultRequestTimeout = 4 * time.Second
	defaultClientVersion  = "2.0"
)

// Client represents the middleman for executing JSON RPC calls
// to remote NEO RPC nodes.
type Client struct {
	cli      *http.Client
	endpoint *url.URL
	ctx      context.Context
	opts     Options
	requestF func(*request.Raw) (*response.Raw, error)
	wifMu    *sync.Mutex
	wif      *keys.WIF
}

// Options defines options for the RPC client.
// All values are optional. If any duration is not specified
// a default of 4 seconds will be used.
type Options struct {
	// Balancer is an implementation of request.BalanceGetter interface,
	// if not set then the default Client's implementation will be used, but
	// it relies on server support for `getunspents` RPC call which is
	// standard for neo-go, but only implemented as a plugin for C# node. So
	// you can override it here to use NeoScanServer for example.
	Balancer request.BalanceGetter

	// Cert is a client-side certificate, it doesn't work at the moment along
	// with the other two options below.
	Cert           string
	Key            string
	CACert         string
	DialTimeout    time.Duration
	RequestTimeout time.Duration
}

// New returns a new Client ready to use.
func New(ctx context.Context, endpoint string, opts Options) (*Client, error) {
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	if opts.DialTimeout <= 0 {
		opts.DialTimeout = defaultDialTimeout
	}

	if opts.RequestTimeout <= 0 {
		opts.RequestTimeout = defaultRequestTimeout
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: opts.DialTimeout,
			}).DialContext,
		},
		Timeout: opts.RequestTimeout,
	}

	// TODO(@antdm): Enable SSL.
	if opts.Cert != "" && opts.Key != "" {
	}

	cl := &Client{
		ctx:      ctx,
		cli:      httpClient,
		wifMu:    new(sync.Mutex),
		endpoint: url,
	}
	if opts.Balancer == nil {
		opts.Balancer = cl
	}
	cl.opts = opts
	cl.requestF = cl.makeHTTPRequest
	return cl, nil
}

// WIF returns WIF structure associated with the client.
func (c *Client) WIF() keys.WIF {
	c.wifMu.Lock()
	defer c.wifMu.Unlock()
	return keys.WIF{
		Version:    c.wif.Version,
		Compressed: c.wif.Compressed,
		PrivateKey: c.wif.PrivateKey,
		S:          c.wif.S,
	}
}

// SetWIF decodes given WIF and adds some wallet
// data to client. Useful for RPC calls that require an open wallet.
func (c *Client) SetWIF(wif string) error {
	c.wifMu.Lock()
	defer c.wifMu.Unlock()
	decodedWif, err := keys.WIFDecode(wif, 0x00)
	if err != nil {
		return errors.Wrap(err, "Failed to decode WIF; failed to add WIF to client ")
	}
	c.wif = decodedWif
	return nil
}

// CalculateInputs implements request.BalanceGetter interface and returns inputs
// array for the specified amount of given asset belonging to specified address.
// This implementation uses GetUnspents JSON-RPC call internally, so make sure
// your RPC server supports that.
func (c *Client) CalculateInputs(address string, asset util.Uint256, cost util.Fixed8) ([]transaction.Input, util.Fixed8, error) {
	var utxos state.UnspentBalances

	resp, err := c.GetUnspents(address)
	if err != nil {
		return nil, util.Fixed8(0), errors.Wrapf(err, "cannot get balance for address %v", address)
	}
	for _, ubi := range resp.Balance {
		if asset.Equals(ubi.AssetHash) {
			utxos = ubi.Unspents
			break
		}
	}
	return unspentsToInputs(utxos, cost)

}

func (c *Client) performRequest(method string, p request.RawParams, v interface{}) error {
	var r = request.Raw{
		JSONRPC:   request.JSONRPCVersion,
		Method:    method,
		RawParams: p.Values,
		ID:        1,
	}

	raw, err := c.requestF(&r)

	if raw != nil && raw.Error != nil {
		return raw.Error
	} else if err != nil {
		return err
	} else if raw == nil || raw.Result == nil {
		return errors.New("no result returned")
	}
	return json.Unmarshal(raw.Result, v)
}

func (c *Client) makeHTTPRequest(r *request.Raw) (*response.Raw, error) {
	var (
		buf = new(bytes.Buffer)
		raw = new(response.Raw)
	)

	if err := json.NewEncoder(buf).Encode(r); err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.endpoint.String(), buf)
	if err != nil {
		return nil, err
	}
	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// The node might send us proper JSON anyway, so look there first and if
	// it parses, then it has more relevant data than HTTP error code.
	err = json.NewDecoder(resp.Body).Decode(raw)
	if err != nil {
		if resp.StatusCode != http.StatusOK {
			err = fmt.Errorf("HTTP %d/%s", resp.StatusCode, http.StatusText(resp.StatusCode))
		} else {
			err = errors.Wrap(err, "JSON decoding")
		}
	}
	if err != nil {
		return nil, err
	}
	return raw, nil
}

// Ping attempts to create a connection to the endpoint.
// and returns an error if there is one.
func (c *Client) Ping() error {
	conn, err := net.DialTimeout("tcp", c.endpoint.Host, defaultDialTimeout)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}
