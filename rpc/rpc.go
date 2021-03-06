package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperchain/gosdk/account"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hyperchain/gosdk/common"
	"github.com/terasum/viper"
)

const (
	// TRANSACTION type
	TRANSACTION = "tx_"
	// CONTRACT type
	CONTRACT = "contract_"
	// BLOCK type
	BLOCK = "block_"
	// ACCOUNT type
	ACCOUNT = "account_"
	// NODE type
	NODE = "node_"
	// CERT type
	CERT = "cert_"
	// SUB type
	SUB = "sub_"
	// ARCHIVE type
	ARCHIVE = "archive_"
	// MQ type
	MQ = "mq_"
	// RADAR type
	RADAR = "radar_"
	// CONFIG type
	CONFIG = "config_"
	// FILE type
	FILE = "fm_"
	// AUTH type
	AUTH = "auth_"
	// SIMULATE type
	SIMULATE = "simulate_"

	DefaultNamespace          = "global"
	DefaultResendTime         = 10
	DefaultFirstPollInterval  = 100
	DefaultFirstPollTime      = 10
	DefaultSecondPollInterval = 1000
	DefaultSecondPollTime     = 10
	DefaultReConnectTime      = 10000
	DefaultTxVersion          = "1.0"
)

var (
	logger    = common.GetLogger("rpc")
	once      = sync.Once{}
	TxVersion = DefaultTxVersion
)

// RPC represents rpc apis
type RPC struct {
	hrm                httpRequestManager
	namespace          string
	resTime            int64
	firstPollInterval  int64
	firstPollTime      int64
	secondPollInterval int64
	secondPollTime     int64
	reConnTime         int64
	txVersion          string
	im                 *inspectorManager
}

type inspectorManager struct {
	enable bool
	key    account.Key
}

func (rpc *RPC) String() string {
	nodes := rpc.hrm.nodes
	var nodeString string
	nodeString += "["
	for i, v := range nodes {
		nodeString += "{\"index\":" + strconv.Itoa(i) + ", \"url:\"" + v.url + "}"
		if i < len(nodes)-1 {
			nodeString += ", "
		}
	}
	nodeString += "]"
	return "\"namespace\":" + rpc.namespace + ", \"nodeUrl\":" + nodeString
}

// NewRPC get a RPC instance with default conf directory path "../conf"
func NewRPC() *RPC {
	return NewRPCWithPath(common.DefaultConfRootPath)
}

type Config struct {
	Namespace     string
	ReConnectTime int64
	JsonRPC       struct {
		Nodes []string
		Ports []string
	}
	WebSocket struct {
		Ports []string
	}
	Polling struct {
		ResendTime            int64
		FirstPollingInterval  int64
		FirstPollingTimes     int64
		SecondPollingInterval int64
		SecondPollingTimes    int64
	}
	Privacy struct {
		SendTcert       bool
		SdkcertPath     string
		SdkcertPrivPath string
		UniquePubPath   string
		UniquePrivPath  string
		Cfca            bool
	}
	Security struct {
		Https       bool
		Tlsca       string
		TlspeerCert string
		TlspeerPriv string
	}
	Log struct {
		LogLevel string
		LogDir   string
	}
	Transport struct {
		MaxIdleConns        int64
		MaxIdleConnsPerHost int64
	}
	Inspector struct {
		Enable         bool
		DefaultAccount string
		AccountType    string
	}
	Tx struct {
		Version string
	}
}

func InitVip(vip *viper.Viper, config *Config) {
	vip.Set(common.NamespaceConf, config.Namespace)
	logger.Debugf("[CONFIG]: %s = %v", common.NamespaceConf, config.Namespace)

	vip.Set(common.ReConnectTime, config.ReConnectTime)
	logger.Debugf("[CONFIG]: %s = %v", common.ReConnectTime, config.ReConnectTime)

	vip.Set(common.JSONRPCNodes, config.JsonRPC.Nodes)
	logger.Debugf("[CONFIG]: %s = %v", common.JSONRPCNodes, config.JsonRPC.Nodes)

	vip.Set(common.JSONRPCPorts, config.JsonRPC.Ports)
	logger.Debugf("[CONFIG]: %s = %v", common.JSONRPCPorts, config.JsonRPC.Ports)

	vip.Set(common.WebSocketPorts, config.WebSocket.Ports)
	logger.Debugf("[CONFIG]: %s = %v", common.WebSocketPorts, config.WebSocket.Ports)

	vip.Set(common.PollingResendTime, config.Polling.ResendTime)
	logger.Debugf("[CONFIG]: %s = %v", common.PollingResendTime, config.Polling.ResendTime)

	vip.Set(common.PollingFirstPollingInterval, config.Polling.FirstPollingInterval)
	logger.Debugf("[CONFIG]: %s = %v", common.PollingFirstPollingInterval, config.Polling.FirstPollingInterval)

	vip.Set(common.PollingFirstPollingTimes, config.Polling.FirstPollingTimes)
	logger.Debugf("[CONFIG]: %s = %v", common.PollingFirstPollingTimes, config.Polling.FirstPollingTimes)

	vip.Set(common.PollingSecondPollingInterval, config.Polling.SecondPollingInterval)
	logger.Debugf("[CONFIG]: %s = %v", common.PollingSecondPollingInterval, config.Polling.SecondPollingInterval)

	vip.Set(common.PollingSecondPollingTimes, config.Polling.SecondPollingTimes)
	logger.Debugf("[CONFIG]: %s = %v", common.PollingSecondPollingTimes, config.Polling.SecondPollingTimes)

	vip.Set(common.PrivacySendTcert, config.Privacy.SendTcert)
	logger.Debugf("[CONFIG]: %s = %v", common.PrivacySendTcert, config.Privacy.SendTcert)

	vip.Set(common.PrivacySDKcertPath, config.Privacy.SdkcertPath)
	logger.Debugf("[CONFIG]: %s = %v", common.PrivacySDKcertPath, config.Privacy.SdkcertPath)

	vip.Set(common.PrivacySDKcertPrivPath, config.Privacy.SdkcertPrivPath)
	logger.Debugf("[CONFIG]: %s = %v", common.PrivacySDKcertPrivPath, config.Privacy.SdkcertPrivPath)

	vip.Set(common.PrivacyUniquePubPath, config.Privacy.UniquePubPath)
	logger.Debugf("[CONFIG]: %s = %v", common.PrivacyUniquePubPath, config.Privacy.UniquePubPath)

	vip.Set(common.PrivacyUniquePrivPath, config.Privacy.UniquePrivPath)
	logger.Debugf("[CONFIG]: %s = %v", common.PrivacyUniquePrivPath, config.Privacy.UniquePrivPath)

	vip.Set(common.PrivacyCfca, config.Privacy.Cfca)
	logger.Debugf("[CONFIG]: %s = %v", common.PrivacyCfca, config.Privacy.Cfca)

	vip.Set(common.SecurityHttps, config.Security.Https)
	logger.Debugf("[CONFIG]: %s = %v", common.SecurityHttps, config.Security.Https)

	vip.Set(common.SecurityTlsca, config.Security.Tlsca)
	logger.Debugf("[CONFIG]: %s = %v", common.SecurityTlsca, config.Security.Tlsca)

	vip.Set(common.SecurityTlspeerCert, config.Security.TlspeerCert)
	logger.Debugf("[CONFIG]: %s = %v", common.SecurityTlspeerCert, config.Security.TlspeerCert)

	vip.Set(common.SecurityTlspeerPriv, config.Security.TlspeerPriv)
	logger.Debugf("[CONFIG]: %s = %v", common.SecurityTlspeerPriv, config.Security.TlspeerPriv)

	vip.Set(common.LogOutputLevel, config.Log.LogLevel)
	logger.Debugf("[CONFIG]: %s = %v", common.LogOutputLevel, config.Log.LogLevel)

	vip.Set(common.LogDir, config.Log.LogDir)
	logger.Debugf("[CONFIG]: %s = %v", common.LogDir, config.Log.LogDir)

	vip.Set(common.MaxIdleConns, config.Transport.MaxIdleConns)
	logger.Debugf("[CONFIG]: %s = %v", common.MaxIdleConns, config.Transport.MaxIdleConns)
	vip.Set(common.MaxIdleConnsPerHost, config.Transport.MaxIdleConnsPerHost)
	logger.Debugf("[CONFIG]: %s = %v", common.MaxIdleConnsPerHost, config.Transport.MaxIdleConnsPerHost)

	vip.Set(common.InspectorEnable, config.Inspector.Enable)
	logger.Debugf("[CONFIG]: %s = %v", common.InspectorEnable, config.Inspector.Enable)
	vip.Set(common.InspectorAccountPath, config.Inspector.DefaultAccount)
	logger.Debugf("[CONFIG]: %s = %v", common.InspectorAccountPath, config.Inspector.DefaultAccount)
	vip.Set(common.InspectorAccountType, config.Inspector.AccountType)
	logger.Debugf("[CONFIG]: %s = %v", common.InspectorAccountType, config.Inspector.AccountType)

	vip.Set(common.TxVersion, config.Tx.Version)
	logger.Debugf("[CONFIG]: %s = %v", common.TxVersion, config.Tx.Version)
}

func NewRPCWithConfig(config *Config) *RPC {
	vip := viper.New()
	InitVip(vip, config)
	common.InitLog(vip)

	im := newInspectorManager2(vip)
	version := vip.GetString(common.TxVersion)
	httpRequestManager := newHTTPRequestManager2(vip, version)
	fmt.Println(httpRequestManager)
	rpc := &RPC{
		hrm:                *httpRequestManager,
		namespace:          config.Namespace,
		resTime:            config.Polling.ResendTime,
		firstPollInterval:  config.Polling.FirstPollingInterval,
		firstPollTime:      config.Polling.FirstPollingTimes,
		secondPollInterval: config.Polling.SecondPollingInterval,
		secondPollTime:     config.Polling.SecondPollingTimes,
		reConnTime:         config.ReConnectTime,
		im:                 im,
	}
	txVersion, err := rpc.GetTxVersion()
	if err != nil {
		logger.Info("use config txVersion, for", err.Error())
		txVersion = version
	}
	TxVersion = txVersion
	rpc.txVersion = txVersion
	rpc.hrm.txVersion = txVersion
	logger.Info("set TxVersion to " + TxVersion)
	return rpc
}

// NewRPCWithPath get a RPC instance with user defined root conf directory path
// the default conf root file structure should like this:
//
//      conf
//		????????? certs
//		??????? ????????? ecert.cert
//		??????? ????????? ecert.priv
//		??????? ????????? sdkcert.cert
//		??????? ????????? sdkcert.priv
//		??????? ????????? tls
//		??????? ??????? ????????? tls_peer.cert
//		??????? ??????? ????????? tls_peer.priv
//		??????? ??????? ????????? tlsca.ca
//		??????? ????????? unique.priv
//		??????? ????????? unique.pub
//		????????? hpc.toml
func NewRPCWithPath(confRootPath string) *RPC {
	vip := viper.New()
	vip.SetConfigFile(filepath.Join(confRootPath, common.DefaultConfRelPath))
	err := vip.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("read conf from %s error", filepath.Join(confRootPath, common.DefaultConfRelPath)))
	}

	common.InitLog(vip)
	namespace := vip.GetString(common.NamespaceConf)
	logger.Debugf("[CONFIG]: %s = %v", common.NamespaceConf, namespace)

	resTime := vip.GetInt64(common.PollingResendTime)
	logger.Debugf("[CONFIG]: %s = %v", common.PollingResendTime, resTime)

	firstPollInterval := vip.GetInt64(common.PollingFirstPollingInterval)
	logger.Debugf("[CONFIG]: %s = %v", common.PollingFirstPollingInterval, firstPollInterval)

	firstPollTime := vip.GetInt64(common.PollingFirstPollingTimes)
	logger.Debugf("[CONFIG]: %s = %v", common.PollingFirstPollingTimes, firstPollTime)

	secondPollInterval := vip.GetInt64(common.PollingSecondPollingInterval)
	logger.Debugf("[CONFIG]: %s = %v", common.PollingSecondPollingInterval, secondPollInterval)

	secondPollTime := vip.GetInt64(common.PollingSecondPollingTimes)
	logger.Debugf("[CONFIG]: %s = %v", common.PollingSecondPollingTimes, secondPollTime)
	reConnTime := vip.GetInt64(common.ReConnectTime)
	logger.Debugf("[CONFIG]: %s = %v", common.ReConnectTime, reConnTime)

	version := vip.GetString(common.TxVersion)
	logger.Debugf("[CONFIG]: %s = %v", common.TxVersion, version)

	im := newInspectorManager(vip, confRootPath)

	httpRequestManager := newHTTPRequestManager(vip, confRootPath, version)

	rpc := &RPC{
		hrm:                *httpRequestManager,
		namespace:          namespace,
		resTime:            resTime,
		firstPollInterval:  firstPollInterval,
		firstPollTime:      firstPollTime,
		secondPollInterval: secondPollInterval,
		secondPollTime:     secondPollTime,
		reConnTime:         reConnTime,
		im:                 im,
	}
	//rpc := DefaultRPC(httpRequestManager.nodes...)
	//	rpc.im = im
	//	txVersion, err := rpc.GetTxVersion()
	//	if err != nil {
	//		logger.Infof("get txVersion err:%v", err)
	//		txVersion = DefaultTxVersion
	//	}
	//	TxVersion = txVersion
	//
	//	logger.Info("set TxVersion to " + TxVersion)
	txVersion, err := rpc.GetTxVersion()
	if err != nil {
		logger.Info("use config txVersion, for", err.Error())
		txVersion = version
	}
	TxVersion = txVersion
	rpc.txVersion = txVersion
	rpc.hrm.txVersion = txVersion
	logger.Info("set TxVersion to " + TxVersion)
	return rpc
}

func newInspectorManager2(vip *viper.Viper) (im *inspectorManager) {
	var err error
	inspectorEnable := vip.GetBool(common.InspectorEnable)
	logger.Debugf("[CONFIG]: %s = %v", common.InspectorEnable, inspectorEnable)

	im = &inspectorManager{
		enable: inspectorEnable,
	}

	if !inspectorEnable {
		return
	}

	accountPath := vip.GetString(common.InspectorAccountPath)
	logger.Debugf("[CONFIG]: %s = %v", common.InspectorAccountPath, accountPath)

	data := []byte(accountPath)
	//data, err := ioutil.ReadFile(accountPath)
	//if err != nil {
	//	logger.Errorf("read %s:%s err:%v", common.InspectorAccountPath, accountPath, err)
	//	return
	//}

	accountType := vip.GetString(common.InspectorAccountType)
	logger.Debugf("[CONFIG]: %s = %v", common.InspectorAccountType, accountType)

	var key account.Key
	switch accountType {
	case "ecdsa":
		key, err = account.NewAccountFromAccountJSON(string(data), "")
	case "sm2":
		key, err = account.NewAccountSm2FromAccountJSON(string(data), "")
	case "ecdsaPriv":
		key, err = account.NewAccountFromPriv(string(data))
	case "ecdsaPrivR1":
		key, err = account.NewAccountR1FromPriv(string(data))
	case "sm2Priv":
		key, err = account.NewAccountSm2FromPriv(string(data))
	default:
		logger.Errorf("unsupport account type:%s", accountType)
		return
	}
	if err != nil {
		logger.Errorf("new account type %s from %s err:%v", accountType, accountPath, err)
		return
	}
	im.key = key
	return
}

func newInspectorManager(vip *viper.Viper, confRootPath string) (im *inspectorManager) {
	inspectorEnable := vip.GetBool(common.InspectorEnable)
	logger.Debugf("[CONFIG]: %s = %v", common.InspectorEnable, inspectorEnable)

	im = &inspectorManager{
		enable: inspectorEnable,
	}

	if !inspectorEnable {
		return
	}

	accountPath := strings.Join([]string{confRootPath, vip.GetString(common.InspectorAccountPath)}, "/")
	logger.Debugf("[CONFIG]: %s = %v", common.InspectorAccountPath, accountPath)

	data, err := ioutil.ReadFile(accountPath)
	if err != nil {
		logger.Errorf("read %s:%s err:%v", common.InspectorAccountPath, accountPath, err)
		return
	}

	accountType := vip.GetString(common.InspectorAccountType)
	logger.Debugf("[CONFIG]: %s = %v", common.InspectorAccountType, accountType)

	var key account.Key
	switch accountType {
	case "ecdsa":
		key, err = account.NewAccountFromAccountJSON(string(data), "")
	case "sm2":
		key, err = account.NewAccountSm2FromAccountJSON(string(data), "")
	case "ecdsaPriv":
		key, err = account.NewAccountFromPriv(string(data))
	case "ecdsaPrivR1":
		key, err = account.NewAccountR1FromPriv(string(data))
	case "sm2Priv":
		key, err = account.NewAccountSm2FromPriv(string(data))
	default:
		logger.Errorf("unsupport account type:%s", accountType)
		return
	}
	if err != nil {
		logger.Errorf("new account type %s from %s err:%v", accountType, accountPath, err)
		return
	}
	im.key = key
	return
}

// DefaultRPC return a *RPC with some default configs
func DefaultRPC(nodes ...*Node) *RPC {
	rpc := &RPC{
		namespace:          DefaultNamespace,
		resTime:            DefaultResendTime,
		firstPollInterval:  DefaultFirstPollInterval,
		firstPollTime:      DefaultFirstPollTime,
		secondPollInterval: DefaultSecondPollInterval,
		secondPollTime:     DefaultSecondPollTime,
		reConnTime:         DefaultReConnectTime,
		hrm:                *defaultHTTPRequestManager(),
		txVersion:          DefaultTxVersion,
	}
	rpc.hrm.nodes = nodes

	return rpc
}

// Namespace setter
func (rpc *RPC) Namespace(ns string) *RPC {
	rpc.namespace = ns
	return rpc
}

// Close close release goroutine and http connection
func (rpc *RPC) Close() {
	rpc.hrm.client.CloseIdleConnections()
}

// ResendTimes setter
func (rpc *RPC) ResendTimes(resTime int64) *RPC {
	rpc.resTime = resTime
	return rpc
}

// FirstPollInterval setter
func (rpc *RPC) FirstPollInterval(fpi int64) *RPC {
	rpc.firstPollInterval = fpi
	return rpc
}

// FirstPollTime setter
func (rpc *RPC) FirstPollTime(fpt int64) *RPC {
	rpc.firstPollTime = fpt
	return rpc
}

// SecondPollInterval setter
func (rpc *RPC) SecondPollInterval(spi int64) *RPC {
	rpc.secondPollInterval = spi
	return rpc
}

// SecondPollTime setter
func (rpc *RPC) SecondPollTime(spt int64) *RPC {
	rpc.secondPollTime = spt
	return rpc
}

// ReConnTime setter
func (rpc *RPC) ReConnTime(rct int64) *RPC {
	rpc.reConnTime = rct
	return rpc
}

// Https use sets the https related options
func (rpc *RPC) Https(tlscaPath, tlspeerCertPath, tlspeerPrivPath string) *RPC {
	vip := viper.New()
	vip.Set(common.SecurityHttps, true)
	vip.Set(common.SecurityTlsca, tlscaPath)
	vip.Set(common.SecurityTlspeerCert, tlspeerCertPath)
	vip.Set(common.SecurityTlspeerPriv, tlspeerPrivPath)

	rpc.hrm.client = newHTTPClient(vip, ".")
	rpc.hrm.isHTTP = true

	for i := 0; i < len(rpc.hrm.nodes); i++ {
		rpc.hrm.nodes[i].url = "https://" + strings.Split(rpc.hrm.nodes[i].url, "//")[1]
	}

	return rpc
}

func (rpc *RPC) AddNode(url, rpcPort, wsPort string) *RPC {
	rpc.hrm.nodes = append(rpc.hrm.nodes, newNode(url, rpcPort, wsPort, rpc.hrm.isHTTP))

	return rpc
}

func (rpc *RPC) Tcert(cfca bool, sdkcertPath, sdkcertPrivPath, uniquePubPath, uniquePrivPath string) *RPC {
	vip := viper.New()
	vip.Set(common.PrivacyCfca, cfca)
	vip.Set(common.PrivacySendTcert, true)
	vip.Set(common.PrivacySDKcertPath, sdkcertPath)
	vip.Set(common.PrivacySDKcertPrivPath, sdkcertPrivPath)
	vip.Set(common.PrivacyUniquePubPath, uniquePubPath)
	vip.Set(common.PrivacyUniquePrivPath, uniquePrivPath)

	rpc.hrm.tcm = NewTCertManager(vip, ".")

	return rpc
}

// BindNodes generate a new RPC instance that bind with given indexes
func (rpc *RPC) BindNodes(nodeIndexes ...int) (*RPC, error) {
	if len(nodeIndexes) == 0 {
		return rpc, nil
	}
	proxy := *rpc
	proxy.hrm.nodes = make([]*Node, len(nodeIndexes))
	proxy.hrm.nodeIndex = 0

	limit := len(rpc.hrm.nodes)
	for i := 0; i < len(nodeIndexes); i++ {
		if nodeIndexes[i] > limit {
			return nil, fmt.Errorf("nodeIndex %d is out of range", i)
		}
		proxy.hrm.nodes[i] = rpc.hrm.nodes[nodeIndexes[i]-1]
	}
	return &proxy, nil
}

// package method name and params to JsonRequest
func (rpc *RPC) jsonRPC(method string, params ...interface{}) *JSONRequest {
	req := &JSONRequest{
		Method:    method,
		Version:   JSONRPCVersion,
		ID:        1,
		Namespace: rpc.namespace,
		Params:    params,
	}
	if rpc.im.enable {
		auth := &Authentication{
			Address:   rpc.im.key.GetAddress(),
			Timestamp: time.Now().UnixNano(),
		}
		sig, err := sign(rpc.im.key, authNeedHash(auth), false, false)
		if err != nil {
			logger.Errorf("sign auth fail")
		}
		auth.Signature = sig
		req.Auth = auth
	}
	return req
}

func authNeedHash(auth *Authentication) string {
	return "address=" + auth.Address.Hex() +
		"&timestamp=0x" + strconv.FormatInt(auth.Timestamp, 16)
}

// call is a function to get response result commodiously
func (rpc *RPC) call(method string, params ...interface{}) (json.RawMessage, StdError) {
	req := rpc.jsonRPC(method, params...)
	return rpc.callWithReq(req)
}

// callWithReq is a function to get response origin data
func (rpc *RPC) callWithReq(req *JSONRequest) (json.RawMessage, StdError) {
	body, sysErr := json.Marshal(req)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	data, err := rpc.hrm.SyncRequest(body)
	if err != nil {
		return nil, err
	}

	var resp *JSONResponse
	if sysErr = json.Unmarshal(data, &resp); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	if resp.Code != SuccessCode {
		return nil, NewServerError(resp.Code, resp.Message)
	}

	return resp.Result, nil
}

// callWithSpecificUrl is a function to get response form specific url
func (rpc *RPC) callWithSpecificURL(method string, url string, params ...interface{}) (json.RawMessage, StdError) {
	req := rpc.jsonRPC(method, params...)

	body, sysErr := json.Marshal(req)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	data, err := rpc.hrm.SyncRequestSpecificURL(body, url, GENERAL, nil, nil)
	if err != nil {
		return nil, err
	}

	var resp *JSONResponse
	if sysErr = json.Unmarshal(data, &resp); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	if resp.Code != SuccessCode {
		return nil, NewServerError(resp.Code, resp.Message)
	}

	return resp.Result, nil
}

// Call call and get tx receipt directly without polling
func (rpc *RPC) Call(method string, param interface{}) (*TxReceipt, StdError) {
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}
	var receipt TxReceipt
	if sysErr := json.Unmarshal(data, &receipt); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return &receipt, nil
}

// CallByPolling call and get tx receipt by polling
func (rpc *RPC) CallByPolling(method string, param interface{}, isPrivateTx bool) (*TxReceipt, StdError) {
	var (
		req    *JSONRequest
		data   json.RawMessage
		hash   string
		err    StdError
		sysErr error
	)
	// if simulate is false, transaction need to resend
	req = rpc.jsonRPC(method, param)
	for i := int64(0); i < rpc.resTime; i++ {
		if data, err = rpc.callWithReq(req); err != nil {
			return nil, err
		} else {
			if sysErr = json.Unmarshal(data, &hash); sysErr != nil {
				return nil, NewSystemError(sysErr)
			}
			txReceipt, innErr, success := rpc.GetTxReceiptByPolling(hash, isPrivateTx)
			err = innErr
			if success {
				return txReceipt, err
			}
			continue
		}
		//if code is -9999 -32001 and -32006, we should sleep then resend
		time.Sleep(time.Millisecond * time.Duration(rpc.firstPollInterval+rpc.secondPollInterval))
	}
	return nil, NewRequestTimeoutError(errors.New("request time out"))
}

func (rpc *RPC) GetTxVersion() (string, StdError) {
	method := TRANSACTION + "getTransactionsVersion"
	data, err := rpc.call(method)
	if err != nil {
		return "", err
	}
	var txVersion string
	if sysErr := json.Unmarshal(data, &txVersion); sysErr != nil {
		return "", NewSystemError(sysErr)
	}
	return txVersion, nil
}

// GetTxReceiptByPolling get tx receipt by polling
func (rpc *RPC) GetTxReceiptByPolling(txHash string, isPrivateTx bool) (*TxReceipt, StdError, bool) {
	var (
		err     StdError
		receipt *TxReceipt
	)
	txHash = chPrefix(txHash)

	for j := int64(0); j < rpc.firstPollTime; j++ {
		receipt, err = rpc.GetTxReceipt(txHash, isPrivateTx)
		if err != nil {
			if err.Code() == BalanceInsufficientCode {
				return nil, err, true
			} else if err.Code() != DataNotExistCode && err.Code() != SystemBusyCode {
				return nil, err, true
			}
			time.Sleep(time.Millisecond * time.Duration(rpc.firstPollInterval))
		} else {
			return receipt, nil, true
		}
	}
	for j := int64(0); j < rpc.secondPollTime; j++ {
		receipt, err = rpc.GetTxReceipt(txHash, isPrivateTx)
		if err != nil {
			if err.Code() == BalanceInsufficientCode {
				return nil, err, true
			} else if err.Code() != DataNotExistCode && err.Code() != SystemBusyCode {
				return nil, err, true
			}
			time.Sleep(time.Millisecond * time.Duration(rpc.secondPollInterval))
		} else {
			return receipt, nil, true
		}
	}
	return nil, NewGetResponseError(errors.New("polling failure")), false
}

/*---------------------------------- node ----------------------------------*/

// GetNodes ???????????????????????????
func (rpc *RPC) GetNodes() ([]NodeInfo, StdError) {
	data, err := rpc.call(NODE + "getNodes")
	if err != nil {
		return nil, err
	}
	var nodeInfo []NodeInfo
	if sysErr := json.Unmarshal(data, &nodeInfo); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	return nodeInfo, nil
}

// GetNodeHash ??????????????????hash
func (rpc *RPC) GetNodeHash() (string, StdError) {
	data, err := rpc.call(NODE + "getNodeHash")
	if err != nil {
		return "", err
	}
	hash := []byte(data)
	return string(hash), nil
}

// GetNodeHashByID ?????????????????????hash
func (rpc *RPC) GetNodeHashByID(id int) (string, StdError) {
	url := rpc.hrm.nodes[id-1].url
	data, err := rpc.callWithSpecificURL(NODE+"getNodeHash", url)
	if err != nil {
		return "", err
	}

	var hash string
	if sysErr := json.Unmarshal(data, &hash); sysErr != nil {
		return "", NewSystemError(sysErr)
	}
	return hash, nil
}

// DeleteNodeVP ??????VP??????
func (rpc *RPC) DeleteNodeVP(hash string) (bool, StdError) {
	method := NODE + "deleteVP"
	param := newMapParam("nodehash", hash)
	_, err := rpc.call(method, param.Serialize())
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteNodeNVP ??????NVP??????
func (rpc *RPC) DeleteNodeNVP(hash string) (bool, StdError) {
	method := NODE + "deleteNVP"
	param := newMapParam("nodehash", hash)
	_, err := rpc.call(method, param.Serialize())
	if err != nil {
		return false, err
	}
	return true, nil
}

// DisconnectNodeVP  NVP?????????VP???????????????
func (rpc *RPC) DisconnectNodeVP(hash string) (bool, StdError) {
	method := NODE + "disconnectVP"
	param := newMapParam("nodehash", hash)
	_, err := rpc.call(method, param.Serialize())
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetNodeStates ????????????????????????
func (rpc *RPC) GetNodeStates() ([]NodeStateInfo, StdError) {
	method := NODE + "getNodeStates"
	data, err := rpc.call(method)
	if err != nil {
		return nil, err
	}

	var list []NodeStateInfo
	if sysErr := json.Unmarshal(data, &list); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return list, nil
}

/*---------------------------------- block ----------------------------------*/

// GetLatestBlock returns information about the latest block
func (rpc *RPC) GetLatestBlock() (*Block, StdError) {
	method := BLOCK + "latestBlock"
	data, stdErr := rpc.call(method)
	if stdErr != nil {
		return nil, stdErr
	}

	blockRaw := BlockRaw{}

	sysErr := json.Unmarshal(data, &blockRaw)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	block, stdErr := blockRaw.ToBlock()
	if stdErr != nil {
		return nil, NewSystemError(sysErr)
	}

	return block, nil
}

// Deprecated
// GetBlocks returns a list of blocks from start block number to end block number
// isPlain indicates if the result includes transaction information. if false, includes, otherwise not.
func (rpc *RPC) GetBlocks(from, to uint64, isPlain bool) ([]*Block, StdError) {
	if from == 0 || to == 0 || to < from {
		return nil, NewSystemError(errors.New("to and from should be non-zero integer and to should no more than from"))
	}

	method := BLOCK + "getBlocks"

	mp := newMapParam("from", from)
	mp.addKV("to", to)
	mp.addKV("isPlain", isPlain)

	data, stdErr := rpc.call(method, mp.Serialize())
	if stdErr != nil {
		return nil, stdErr
	}

	var blockRaws []BlockRaw

	sysErr := json.Unmarshal(data, &blockRaws)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	blocks := make([]*Block, 0, len(blockRaws))

	for _, v := range blockRaws {
		block, stdErr := v.ToBlock()
		if stdErr != nil {
			return nil, stdErr
		}

		blocks = append(blocks, block)
	}

	return blocks, nil

}

func (rpc *RPC) GetBlocksWithLimit(from, to uint64, isPlain bool, metadata *Metadata) (*PageResult, StdError) {
	if from == 0 || to == 0 || to < from {
		return nil, NewSystemError(errors.New("to and from should be non-zero integer and to should no more than from"))
	}

	method := BLOCK + "getBlocksWithLimit"

	mp := newMapParam("from", from)
	mp.addKV("to", to)
	mp.addKV("isPlain", isPlain)
	mp.addKV("matadata", metadata)

	data, stdErr := rpc.call(method, mp.Serialize())
	if stdErr != nil {
		return nil, stdErr
	}

	var pageResult *PageResult
	sysErr := json.Unmarshal(data, &pageResult)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	return pageResult, nil
}

// GetBlockByHash returns information about a block by hash.
// If the param isPlain value is true, it returns block excluding transactions. If false,
// it returns block including transactions.
func (rpc *RPC) GetBlockByHash(blockHash string, isPlain bool) (*Block, StdError) {
	method := BLOCK + "getBlockByHash"

	data, stdErr := rpc.call(method, blockHash, isPlain)
	if stdErr != nil {
		return nil, stdErr
	}

	blockRaw := BlockRaw{}
	if sysErr := json.Unmarshal(data, &blockRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	block, stdErr := blockRaw.ToBlock()
	if stdErr != nil {
		return nil, stdErr
	}

	return block, nil
}

// GetBatchBlocksByHash returns a list of blocks by a list of specific block hash.
func (rpc *RPC) GetBatchBlocksByHash(blockHashes []string, isPlain bool) ([]*Block, StdError) {
	method := BLOCK + "getBatchBlocksByHash"

	mp := newMapParam("hashes", blockHashes)
	mp.addKV("isPlain", isPlain)

	data, stdErr := rpc.call(method, mp.Serialize())
	if stdErr != nil {
		return nil, stdErr
	}

	var blockRaws []BlockRaw

	sysErr := json.Unmarshal(data, &blockRaws)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	blocks := make([]*Block, 0, len(blockRaws))

	for _, v := range blockRaws {
		block, stdErr := v.ToBlock()
		if stdErr != nil {
			return nil, stdErr
		}

		blocks = append(blocks, block)
	}

	return blocks, nil
}

// GetBlockByNumber returns information about a block by number. If the param isPlain
// value is true, it returns block excluding transactions. If false, it returns block
// including transactions.
// blockNum can use `latest`, means get latest block
func (rpc *RPC) GetBlockByNumber(blockNum interface{}, isPlain bool) (*Block, StdError) {
	method := BLOCK + "getBlockByNumber"

	data, stdErr := rpc.call(method, blockNum, isPlain)
	if stdErr != nil {
		return nil, stdErr
	}

	var blockRaw BlockRaw

	sysErr := json.Unmarshal(data, &blockRaw)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	block, stdErr := blockRaw.ToBlock()
	if stdErr != nil {
		return nil, stdErr
	}

	return block, nil
}

// GetBatchBlocksByNumber returns a list of blocks by a list of specific block number.
func (rpc *RPC) GetBatchBlocksByNumber(blockNums []uint64, isPlain bool) ([]*Block, StdError) {
	method := BLOCK + "getBatchBlocksByNumber"

	mp := newMapParam("numbers", blockNums)
	mp.addKV("isPlain", isPlain)

	data, stdErr := rpc.call(method, mp.Serialize())
	if stdErr != nil {
		return nil, stdErr
	}

	var blockRaws []BlockRaw

	sysErr := json.Unmarshal(data, &blockRaws)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	blocks := make([]*Block, 0, len(blockRaws))

	for _, v := range blockRaws {
		block, stdErr := v.ToBlock()
		if stdErr != nil {
			return nil, stdErr
		}

		blocks = append(blocks, block)
	}

	return blocks, nil
}

// GetAvgGenTimeByBlockNum calculates the average generation time of all blocks
// for the given block number.
func (rpc *RPC) GetAvgGenTimeByBlockNum(from, to uint64) (int64, StdError) {
	if from == 0 || to == 0 || to < from {
		return -1, NewSystemError(errors.New("to and from should be non-zero integer and to should no more than from"))
	}

	method := BLOCK + "getAvgGenerateTimeByBlockNumber"

	mp := newMapParam("from", from)
	mp.addKV("to", to)

	data, stdErr := rpc.call(method, mp.Serialize())
	if stdErr != nil {
		return -1, stdErr
	}

	str := strings.Replace(string(data), "\"", "", 2)

	if strings.Index(str, "0x") == 0 || strings.Index(str, "-0x") == 0 {
		str = strings.Replace(str, "0x", "", 1)
	}

	avgTime, sysErr := strconv.ParseInt(str, 16, 64)
	if sysErr != nil {
		return -1, NewSystemError(sysErr)
	}

	return avgTime, nil
}

// GetBlocksByTime returns the number of blocks, starting block and ending block
// at specific time periods.
// startTime and endTime are timestamps
func (rpc *RPC) GetBlocksByTime(startTime, endTime uint64) (*BlockInterval, StdError) {
	if endTime < startTime {
		return nil, NewSystemError(errors.New("startTime should less than endTime"))
	}

	method := BLOCK + "getBlocksByTime"

	mp := newMapParam("startTime", startTime)
	mp.addKV("endTime", endTime)

	data, stdErr := rpc.call(method, mp.Serialize())
	if stdErr != nil {
		return nil, stdErr
	}

	var blockNumRaw BlockIntervalRaw

	sysErr := json.Unmarshal(data, &blockNumRaw)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	blockNum, stdErr := blockNumRaw.ToBlockInterval()
	if stdErr != nil {
		return nil, stdErr
	}

	return blockNum, nil
}

// QueryTPS queries the block generation speed and tps within a given time range.
func (rpc *RPC) QueryTPS(startTime, endTime uint64) (*TPSInfo, StdError) {
	if endTime < startTime {
		return nil, NewSystemError(errors.New("startTime should less than endTime"))
	}

	method := BLOCK + "queryTPS"

	mp := newMapParam("startTime", startTime)
	mp.addKV("endTime", endTime)

	data, stdErr := rpc.call(method, mp.Serialize())
	if stdErr != nil {
		return nil, stdErr
	}

	items := strings.Split(string(data), ";")

	startTimeStr := items[0][12:]
	endTimeStr := items[1][9:]
	totalBlock, sysErr := strconv.ParseUint(strings.Split(items[2], ":")[1], 0, 64)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	blockPerSec, sysErr := strconv.ParseFloat(strings.Split(items[3], ":")[1], 64)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	tps, sysErr := strconv.ParseFloat(strings.Split(strings.Trim(items[4], "\""), ":")[1], 64)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	return &TPSInfo{
		StartTime:     startTimeStr,
		EndTime:       endTimeStr,
		TotalBlockNum: totalBlock,
		BlocksPerSec:  blockPerSec,
		Tps:           tps,
	}, nil
}

// GetGenesisBlock returns current genesis block number.
// result is hex string
func (rpc *RPC) GetGenesisBlock() (string, StdError) {
	method := BLOCK + "getGenesisBlock"

	data, stdErr := rpc.call(method)
	if stdErr != nil {
		return "", stdErr
	}

	var result string
	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return "", NewSystemError(sysErr)
	}

	return result, nil
}

// GetChainHeight returns the current chain height.
// result is hex string
func (rpc *RPC) GetChainHeight() (string, StdError) {
	method := BLOCK + "getChainHeight"

	data, stdErr := rpc.call(method)
	if stdErr != nil {
		return "", stdErr
	}

	var result string
	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return "", NewSystemError(sysErr)
	}

	return result, nil
}

/*---------------------------------- transaction ----------------------------------*/

// Deprecated
// GetTransactionsByBlkNum ???????????????????????????????????????
func (rpc *RPC) GetTransactionsByBlkNum(start, end uint64) ([]TransactionInfo, StdError) {
	qtr := &QueryTxRange{
		From: start,
		To:   end,
	}
	method := TRANSACTION + "getTransactions"
	param := qtr.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var txsRaw []TransactionRaw
	if sysErr := json.Unmarshal(data, &txsRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	txs := make([]TransactionInfo, 0, len(txsRaw))
	for _, txRaw := range txsRaw {
		t, err := txRaw.ToTransaction()
		if err != nil {
			return nil, err
		}
		txs = append(txs, *t)
	}
	return txs, nil
}

func (rpc *RPC) GetTransactionsByBlkNumWithLimit(start, end uint64, metadata *Metadata) (*PageResult, StdError) {
	qtr := &QueryTxRange{
		From:     start,
		To:       end,
		metadata: metadata,
	}
	method := TRANSACTION + "getTransactionsWithLimit"
	param := qtr.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var pageResult *PageResult
	sysErr := json.Unmarshal(data, &pageResult)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	return pageResult, nil
}

// GetDiscardTx ????????????????????????
func (rpc *RPC) GetDiscardTx() ([]TransactionInfo, StdError) {
	method := TRANSACTION + "getDiscardTransactions"
	data, err := rpc.call(method)
	if err != nil {
		return nil, err
	}

	var txsRaw []TransactionRaw
	if sysErr := json.Unmarshal(data, &txsRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	txs := make([]TransactionInfo, 0, len(txsRaw))
	for _, txRaw := range txsRaw {
		t, err := txRaw.ToTransaction()
		if err != nil {
			return nil, err
		}
		txs = append(txs, *t)
	}
	return txs, nil
}

// GetTransactionByHash ????????????hash????????????
// ??????txHash?????????"0x...."?????????
func (rpc *RPC) GetTransactionByHash(txHash string) (*TransactionInfo, StdError) {
	method := TRANSACTION + "getTransactionByHash"
	param := txHash
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var tx TransactionRaw
	if sysErr := json.Unmarshal(data, &tx); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return tx.ToTransaction()
}

func (rpc *RPC) GetPrivateTransactionByHash(txHash string) (*TransactionInfo, StdError) {
	method := TRANSACTION + "getPrivateTransactionByHash"
	param := txHash
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var tx TransactionRaw
	if sysErr := json.Unmarshal(data, &tx); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return tx.ToTransaction()
}

// GetBatchTxByHash ??????????????????
func (rpc *RPC) GetBatchTxByHash(hashes []string) ([]TransactionInfo, StdError) {
	mp := newMapParam("hashes", hashes)
	method := TRANSACTION + "getBatchTransactions"
	param := mp.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var txsRaw []TransactionRaw
	if sysErr := json.Unmarshal(data, &txsRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	txs := make([]TransactionInfo, 0, len(txsRaw))
	for _, txRaw := range txsRaw {
		t, err := txRaw.ToTransaction()
		if err != nil {
			return nil, err
		}
		txs = append(txs, *t)
	}
	return txs, nil
}

// GetTxByBlkHashAndIdx ????????????hash?????????????????????????????????
func (rpc *RPC) GetTxByBlkHashAndIdx(blkHash string, index uint64) (*TransactionInfo, StdError) {
	method := TRANSACTION + "getTransactionByBlockHashAndIndex"
	data, err := rpc.call(method, blkHash, index)
	if err != nil {
		return nil, err
	}

	var tx TransactionRaw
	if sysErr := json.Unmarshal(data, &tx); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return tx.ToTransaction()
}

// GetTxByBlkNumAndIdx ??????????????????????????????????????????
func (rpc *RPC) GetTxByBlkNumAndIdx(blkNum, index uint64) (*TransactionInfo, StdError) {
	method := TRANSACTION + "getTransactionByBlockNumberAndIndex"
	data, err := rpc.call(method, strconv.FormatUint(blkNum, 10), index)
	if err != nil {
		return nil, err
	}
	var tx TransactionRaw
	if sysErr := json.Unmarshal(data, &tx); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return tx.ToTransaction()
}

// GetTxAvgTimeByBlockNumber ???????????????????????????????????????????????????
func (rpc *RPC) GetTxAvgTimeByBlockNumber(from, to uint64) (uint64, StdError) {
	mp := newMapParam("from", from)
	mp.addKV("to", to)
	method := TRANSACTION + "getTxAvgTimeByBlockNumber"
	param := mp.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return 0, err
	}

	var avgTime string
	if sysErr := json.Unmarshal(data, &avgTime); sysErr != nil {
		return 0, NewSystemError(sysErr)
	}
	result, sysErr := strconv.ParseUint(avgTime, 0, 64)
	if sysErr != nil {
		return 0, NewSystemError(sysErr)
	}
	return result, nil
}

// GetBatchReceipt ??????????????????
func (rpc *RPC) GetBatchReceipt(hashes []string) ([]TxReceipt, StdError) {
	mp := newMapParam("hashes", hashes)
	method := TRANSACTION + "getBatchReceipt"
	param := mp.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var txs []TxReceipt
	if sysErr := json.Unmarshal(data, &txs); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return txs, nil
}

// GetTransactionsCountByTime ??????????????????????????????????????????
func (rpc *RPC) GetTransactionsCountByTime(startTime, endTime uint64) (uint64, StdError) {
	mp := newMapParam("startTime", startTime).addKV("endTime", endTime)
	method := TRANSACTION + "getTransactionsCountByTime"
	param := mp.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return 0, err
	}

	var hexCount string
	if sysError := json.Unmarshal(data, &hexCount); sysError != nil {
		return 0, NewSystemError(err)
	}
	count, sysErr := strconv.ParseUint(hexCount, 0, 64)
	if sysErr != nil {
		return 0, NewSystemError(sysErr)
	}
	return count, nil
}

// GetBlkTxCountByHash ????????????hash????????????????????????
func (rpc *RPC) GetBlkTxCountByHash(blkHash string) (uint64, StdError) {
	method := TRANSACTION + "getBlockTransactionCountByHash"
	param := blkHash
	data, err := rpc.call(method, param)
	if err != nil {
		return 0, err
	}

	var hexCount string
	if sysError := json.Unmarshal(data, &hexCount); sysError != nil {
		return 0, NewSystemError(err)
	}
	count, sysErr := strconv.ParseUint(hexCount, 0, 64)
	if sysErr != nil {
		return 0, NewSystemError(sysErr)
	}
	return count, nil
}

// GetBlkTxCountByNumber ????????????number????????????????????????
func (rpc *RPC) GetBlkTxCountByNumber(blkNum string) (uint64, StdError) {
	method := TRANSACTION + "getBlockTransactionCountByNumber"
	param := blkNum
	data, err := rpc.call(method, param)
	if err != nil {
		return 0, err
	}

	var hexCount string
	if sysError := json.Unmarshal(data, &hexCount); sysError != nil {
		return 0, NewSystemError(err)
	}
	count, sysErr := strconv.ParseUint(hexCount, 0, 64)
	if sysErr != nil {
		return 0, NewSystemError(sysErr)
	}
	return count, nil
}

// GetSignHash ????????????????????????
func (rpc *RPC) GetSignHash(transaction *Transaction) (string, StdError) {
	method := TRANSACTION + "getSignHash"
	param := transaction.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return "", err
	}

	var hash string
	if sysError := json.Unmarshal(data, &hash); sysError != nil {
		return "", NewSystemError(err)
	}
	return hash, nil
}

// GetTxCount ??????????????????????????????
func (rpc *RPC) GetTxCount() (*TransactionsCount, StdError) {
	mehtod := TRANSACTION + "getTransactionsCount"
	data, err := rpc.call(mehtod)
	if err != nil {
		return nil, err
	}

	var txRaw TransactionsCountRaw
	if sysErr := json.Unmarshal(data, &txRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	txCount, sysErr := txRaw.ToTransactionsCount()
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return txCount, nil
}

// GetTxCountByContractAddr ??????????????????????????????????????? txExtra??????????????????????????????
func (rpc *RPC) GetTxCountByContractAddr(from, to uint64, address string, txExtra bool) (*TransactionsCountByContract, StdError) {
	mp := newMapParam("from", from).addKV("to", to).addKV("address", address).addKV("txExtra", txExtra)
	method := TRANSACTION + "getTransactionsCountByContractAddr"
	param := mp.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var countRaw *TransactionsCountByContractRaw
	if sysErr := json.Unmarshal(data, &countRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	count, sysErr := countRaw.ToTransactionsCountByContract()
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return count, nil
}

// GetTxCountByContractName ??????????????????????????????????????? txExtra??????????????????????????????
func (rpc *RPC) GetTxCountByContractName(from, to uint64, name string, txExtra bool) (*TransactionsCountByContract, StdError) {
	mp := newMapParam("from", from).addKV("to", to).addKV("name", name).addKV("txExtra", txExtra)
	method := TRANSACTION + "getTransactionsCountByContractName"
	param := mp.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var countRaw *TransactionsCountByContractRaw
	if sysErr := json.Unmarshal(data, &countRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	count, sysErr := countRaw.ToTransactionsCountByContract()
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return count, nil
}

// GetTransactionsCountByMethodID ?????????????????????????????????by method ID???
func (rpc *RPC) GetTransactionsCountByMethodID(from, to uint64, address string, methodID string) (*TransactionsCountByContract, StdError) {
	mp := newMapParam("from", from).addKV("to", to).addKV("address", address).addKV("methodID", methodID)
	method := TRANSACTION + "getTransactionsCountByMethodID"
	param := mp.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var countRaw *TransactionsCountByContractRaw
	if sysErr := json.Unmarshal(data, &countRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	count, sysErr := countRaw.ToTransactionsCountByContract()
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return count, nil
}

// GetTransactionsCountByMethodIDAndContractName ?????????????????????????????????by method ID and contract name???
func (rpc *RPC) GetTransactionsCountByMethodIDAndContractName(from, to uint64, name string, methodID string) (*TransactionsCountByContract, StdError) {
	mp := newMapParam("from", from).addKV("to", to).addKV("name", name).addKV("methodID", methodID)
	method := TRANSACTION + "getTransactionsCountByMethodIDAndContractName"
	param := mp.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var countRaw *TransactionsCountByContractRaw
	if sysErr := json.Unmarshal(data, &countRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	count, sysErr := countRaw.ToTransactionsCountByContract()
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return count, nil
}

// Deprecated
// GetTxByTime ???????????????????????????????????????
func (rpc *RPC) GetTxByTime(start, end uint64) ([]TransactionInfo, StdError) {
	mp := newMapParam("startTime", start).addKV("endTime", end)
	method := TRANSACTION + "getTransactionsByTime"
	param := mp.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var txsRaw []TransactionRaw
	if sysErr := json.Unmarshal(data, &txsRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	txs := make([]TransactionInfo, 0, len(txsRaw))
	for _, txRaw := range txsRaw {
		t, err := txRaw.ToTransaction()
		if err != nil {
			return nil, err
		}
		txs = append(txs, *t)
	}
	return txs, nil
}

func (rpc *RPC) GetTxByTimeWithLimit(start, end uint64, metadata *Metadata) (*PageTxs, StdError) {
	mp := newMapParam("startTime", start).addKV("endTime", end).addKV("metadata", metadata)
	method := TRANSACTION + "getTransactionsByTimeWithLimit"
	param := mp.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var pageResult *PageTxs
	sysErr := json.Unmarshal(data, &pageResult)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	return pageResult, nil
}

// GetTxByTimeAndContractAddrWithLimit get txs by time and contract address with limit
func (rpc *RPC) GetTxByTimeAndContractAddrWithLimit(start, end uint64, metadata *Metadata, contractAddr string) (*PageTxs, StdError) {
	param := &IntervalTime{
		StartTime: int64(start),
		Endtime:   int64(end),
		Metadata:  metadata,
		Filter: &Filter{
			TxTo: contractAddr,
		},
	}
	return rpc.getTxByTimeWithLimit(param)
}

// GetTxByTimeAndContractNameWithLimit get txs by time and contract name with limit
func (rpc *RPC) GetTxByTimeAndContractNameWithLimit(start, end uint64, metadata *Metadata, contractName string) (*PageTxs, StdError) {
	param := &IntervalTime{
		StartTime: int64(start),
		Endtime:   int64(end),
		Metadata:  metadata,
		Filter: &Filter{
			Name: contractName,
		},
	}
	return rpc.getTxByTimeWithLimit(param)
}

func (rpc *RPC) getTxByTimeWithLimit(param interface{}) (*PageTxs, StdError) {
	method := TRANSACTION + "getTransactionsByTimeWithLimit"
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var pageResult *PageTxs
	sysErr := json.Unmarshal(data, &pageResult)
	if sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	return pageResult, nil
}

// GetDiscardTransactionsByTime ??????????????????????????????????????????
func (rpc *RPC) GetDiscardTransactionsByTime(start, end uint64) ([]TransactionInfo, StdError) {
	mp := newMapParam("startTime", start).addKV("endTime", end)
	method := TRANSACTION + "getDiscardTransactionsByTime"
	param := mp.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var txsRaw []TransactionRaw
	if sysErr := json.Unmarshal(data, &txsRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	txs := make([]TransactionInfo, 0, len(txsRaw))
	for _, txRaw := range txsRaw {
		t, err := txRaw.ToTransaction()
		if err != nil {
			return nil, err
		}
		txs = append(txs, *t)
	}
	return txs, nil
}

// GetNextPageTxs ????????????????????????
func (rpc *RPC) GetNextPageTxs(blkNumber, txIndex, minBlkNumber, maxBlkNumber, separated, pageSize uint64, containCurrent bool, contractAddr string) ([]TransactionInfo, StdError) {
	method := TRANSACTION + "getNextPageTransactions"
	param := &TransactionPageArg{
		BlkNumber:      strconv.FormatUint(blkNumber, 10),
		MaxBlkNumber:   strconv.FormatUint(maxBlkNumber, 10),
		MinBlkNumber:   strconv.FormatUint(minBlkNumber, 10),
		TxIndex:        txIndex,
		Separated:      separated,
		PageSize:       pageSize,
		ContainCurrent: containCurrent,
		Address:        contractAddr,
	}
	return rpc.getPageTxs(method, param)
}

// GetNextPageTxsByName ????????????????????????
func (rpc *RPC) GetNextPageTxsByName(blkNumber, txIndex, minBlkNumber, maxBlkNumber, separated, pageSize uint64, containCurrent bool, contractName string) ([]TransactionInfo, StdError) {
	method := TRANSACTION + "getNextPageTransactions"
	param := &TransactionPageArg{
		BlkNumber:      strconv.FormatUint(blkNumber, 10),
		MaxBlkNumber:   strconv.FormatUint(maxBlkNumber, 10),
		MinBlkNumber:   strconv.FormatUint(minBlkNumber, 10),
		TxIndex:        txIndex,
		Separated:      separated,
		PageSize:       pageSize,
		ContainCurrent: containCurrent,
		CName:          contractName,
	}
	return rpc.getPageTxs(method, param)
}

// GetPrevPageTxs ????????????????????????
func (rpc *RPC) GetPrevPageTxs(blkNumber, txIndex, minBlkNumber, maxBlkNumber, separated, pageSize uint64, containCurrent bool, contractAddr string) ([]TransactionInfo, StdError) {
	method := TRANSACTION + "getPrevPageTransactions"
	param := &TransactionPageArg{
		BlkNumber:      strconv.FormatUint(blkNumber, 10),
		MaxBlkNumber:   strconv.FormatUint(maxBlkNumber, 10),
		MinBlkNumber:   strconv.FormatUint(minBlkNumber, 10),
		TxIndex:        txIndex,
		Separated:      separated,
		PageSize:       pageSize,
		ContainCurrent: containCurrent,
		Address:        contractAddr,
	}
	return rpc.getPageTxs(method, param)
}

// GetPrevPageTxs ????????????????????????
func (rpc *RPC) GetPrevPageTxsByName(blkNumber, txIndex, minBlkNumber, maxBlkNumber, separated, pageSize uint64, containCurrent bool, contractName string) ([]TransactionInfo, StdError) {
	method := TRANSACTION + "getPrevPageTransactions"
	param := &TransactionPageArg{
		BlkNumber:      strconv.FormatUint(blkNumber, 10),
		MaxBlkNumber:   strconv.FormatUint(maxBlkNumber, 10),
		MinBlkNumber:   strconv.FormatUint(minBlkNumber, 10),
		TxIndex:        txIndex,
		Separated:      separated,
		PageSize:       pageSize,
		ContainCurrent: containCurrent,
		CName:          contractName,
	}
	return rpc.getPageTxs(method, param)
}

func (rpc *RPC) getPageTxs(method string, param *TransactionPageArg) ([]TransactionInfo, StdError) {
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var txsRaw []TransactionRaw
	if sysErr := json.Unmarshal(data, &txsRaw); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	txs := make([]TransactionInfo, 0, len(txsRaw))
	for _, txRaw := range txsRaw {
		t, err := txRaw.ToTransaction()
		if err != nil {
			return nil, err
		}
		txs = append(txs, *t)
	}
	return txs, nil
}

// GetTxReceipt ????????????hash??????????????????
// ??????txHash?????????"0x...."?????????
func (rpc *RPC) GetTxReceipt(txHash string, isPrivateTx bool) (*TxReceipt, StdError) {
	var method string
	txHash = chPrefix(txHash)
	if isPrivateTx {
		method = TRANSACTION + "getPrivateTransactionReceipt"
	} else {
		method = TRANSACTION + "getTransactionReceipt"
	}
	param := txHash
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var txr TxReceipt
	if sysErr := json.Unmarshal(data, &txr); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	txr.PrivTxHash = txHash
	return &txr, nil
}

// Deprecated
// SendTx ??????????????????
func (rpc *RPC) SendTx(transaction *Transaction) (*TxReceipt, StdError) {
	transaction.txVersion = rpc.txVersion
	method := TRANSACTION + "sendTransaction"
	param := transaction.Serialize()
	if transaction.simulate {
		return rpc.Call(method, param)
	}
	return rpc.CallByPolling(method, param, transaction.isPrivateTx)
}

// SignAndSendTx ???????????????????????????
func (rpc *RPC) SignAndSendTx(transaction *Transaction, key interface{}) (*TxReceipt, StdError) {
	transaction.txVersion = rpc.txVersion
	transaction.Sign(key)
	method := TRANSACTION + "sendTransaction"
	param := transaction.Serialize()
	if transaction.simulate {
		return rpc.Call(method, param)
	}
	return rpc.CallByPolling(method, param, transaction.isPrivateTx)
}

/*---------------------------------- contract ----------------------------------*/

// CompileContract Compile contract rpc
func (rpc *RPC) CompileContract(code string) (*CompileResult, StdError) {
	data, err := rpc.call(CONTRACT+"compileContract", code)
	if err != nil {
		return nil, err
	}

	var cr CompileResult
	if sysErr := json.Unmarshal(data, &cr); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return &cr, nil
}

func isTxVersion10(txVersion string) bool {
	return strings.Compare(txVersion, "1.0") == 0
}

// Deprecated
// DeployContract Deploy contract rpc
func (rpc *RPC) DeployContract(transaction *Transaction) (*TxReceipt, StdError) {
	var method string
	if transaction.isPrivateTx {
		method = CONTRACT + "deployPrivateContract"
	} else {
		if !isTxVersion10(transaction.getTxVersion()) && transaction.simulate {
			method = SIMULATE + "deployContract"
		} else {
			method = CONTRACT + "deployContract"
		}
	}
	transaction.isDeploy = true
	param := transaction.Serialize()
	if transaction.simulate {
		return rpc.Call(method, param)
	}
	return rpc.CallByPolling(method, param, transaction.isPrivateTx)
}

// SignAndDeployContract Deploy contract rpc
func (rpc *RPC) SignAndDeployContract(transaction *Transaction, key interface{}) (*TxReceipt, StdError) {
	transaction.txVersion = rpc.txVersion
	transaction.Sign(key)
	var method string
	if transaction.isPrivateTx {
		method = CONTRACT + "deployPrivateContract"
	} else {
		if !isTxVersion10(transaction.getTxVersion()) && transaction.simulate {
			method = SIMULATE + "deployContract"
		} else {
			method = CONTRACT + "deployContract"
		}
	}
	transaction.isDeploy = true
	param := transaction.Serialize()
	if transaction.simulate {
		return rpc.Call(method, param)
	}
	return rpc.CallByPolling(method, param, transaction.isPrivateTx)
}

// SignAndInvokeContract invoke contract rpc
func (rpc *RPC) SignAndInvokeContract(transaction *Transaction, key interface{}) (*TxReceipt, StdError) {
	transaction.txVersion = rpc.txVersion
	transaction.Sign(key)
	var method string
	if transaction.isPrivateTx {
		method = CONTRACT + "invokePrivateContract"
	} else {
		if !isTxVersion10(transaction.getTxVersion()) && transaction.simulate {
			method = SIMULATE + "invokeContract"
		} else {
			method = CONTRACT + "invokeContract"
		}
	}
	transaction.isInvoke = true
	param := transaction.Serialize()

	if transaction.simulate {
		return rpc.Call(method, param)
	}
	return rpc.CallByPolling(method, param, transaction.isPrivateTx)
}

// Deprecated
// InvokeContract invoke contract rpc
func (rpc *RPC) InvokeContract(transaction *Transaction) (*TxReceipt, StdError) {
	var method string
	if transaction.isPrivateTx {
		method = CONTRACT + "invokePrivateContract"
	} else {
		if !isTxVersion10(transaction.getTxVersion()) && transaction.simulate {
			method = SIMULATE + "invokeContract"
		} else {
			method = CONTRACT + "invokeContract"
		}
	}
	transaction.isInvoke = true
	param := transaction.Serialize()

	if transaction.simulate {
		return rpc.Call(method, param)
	}
	return rpc.CallByPolling(method, param, transaction.isPrivateTx)
}

// Deprecated
// ManageContractByVote manage contract by vote rpc
func (rpc *RPC) ManageContractByVote(transaction *Transaction) (*TxReceipt, StdError) {
	method := CONTRACT + "manageContractByVote"
	transaction.isInvoke = true
	param := transaction.Serialize()

	if transaction.simulate {
		return rpc.Call(method, param)
	}
	return rpc.CallByPolling(method, param, transaction.isPrivateTx)
}

// ManageContractByVote manage contract by vote rpc
func (rpc *RPC) SignAndManageContractByVote(transaction *Transaction, key interface{}) (*TxReceipt, StdError) {
	transaction.txVersion = rpc.txVersion
	transaction.Sign(key)
	method := CONTRACT + "manageContractByVote"
	transaction.isInvoke = true
	param := transaction.Serialize()

	if transaction.simulate {
		return rpc.Call(method, param)
	}
	return rpc.CallByPolling(method, param, transaction.isPrivateTx)
}

// GetCode ????????????????????????
func (rpc *RPC) GetCode(contractAddress string) (string, StdError) {
	method := CONTRACT + "getCode"
	param := contractAddress
	data, err := rpc.call(method, param)
	if err != nil {
		return "", err
	}

	var code string
	if sysErr := json.Unmarshal(data, &code); sysErr != nil {
		return "", NewSystemError(sysErr)
	}

	return code, nil
}

// GetContractCountByAddr ??????????????????
func (rpc *RPC) GetContractCountByAddr(accountAddress string) (uint64, StdError) {
	method := CONTRACT + "getContractCountByAddr"
	param := accountAddress
	data, err := rpc.call(method, param)
	if err != nil {
		return 0, err
	}

	var hexCount string
	if sysError := json.Unmarshal(data, &hexCount); sysError != nil {
		return 0, NewSystemError(err)
	}
	count, sysErr := strconv.ParseUint(hexCount, 0, 64)
	if sysErr != nil {
		return 0, NewSystemError(sysErr)
	}
	return count, nil
}

// EncryptoMessage ?????????????????????????????????????????????????????????
func (rpc *RPC) EncryptoMessage(balance, amount uint64, invalidHmValue string) (*BalanceAndAmount, StdError) {
	mp := newMapParam("balance", balance).addKV("amount", amount).addKV("invalidHmValue", invalidHmValue)
	method := CONTRACT + "encryptoMessage"
	param := mp.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var balanceAndAmount *BalanceAndAmount
	if sysError := json.Unmarshal(data, &balanceAndAmount); sysError != nil {
		return nil, NewSystemError(err)
	}

	return balanceAndAmount, nil
}

// CheckHmValue ????????????????????????????????????????????????????????????
func (rpc *RPC) CheckHmValue(rawValue []uint64, encryValue []string, invalidHmValue string) (*ValidResult, StdError) {
	mp := newMapParam("rawValue", rawValue).addKV("encryValue", encryValue).addKV("invalidHmValue", invalidHmValue)
	method := CONTRACT + "checkHmValue"
	param := mp.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var validResutl *ValidResult
	if sysError := json.Unmarshal(data, &validResutl); sysError != nil {
		return nil, NewSystemError(err)
	}

	return validResutl, nil
}

// Deprecated
// MaintainContract ???????????? opcode
// 1.????????????
// 2.??????
// 3.??????
func (rpc *RPC) MaintainContract(transaction *Transaction) (*TxReceipt, StdError) {
	var method string
	if !isTxVersion10(transaction.getTxVersion()) && transaction.simulate {
		method = SIMULATE + "maintainContract"
	} else {
		method = CONTRACT + "maintainContract"
	}
	transaction.isMaintain = true
	param := transaction.Serialize()
	if transaction.simulate {
		return rpc.Call(method, param)
	}
	return rpc.CallByPolling(method, param, transaction.isPrivateTx)
}

// SignAndMaintainContract ???????????? opcode
// 1.????????????
// 2.??????
// 3.??????
func (rpc *RPC) SignAndMaintainContract(transaction *Transaction, key interface{}) (*TxReceipt, StdError) {
	transaction.txVersion = rpc.txVersion
	transaction.Sign(key)
	var method string
	if !isTxVersion10(transaction.getTxVersion()) && transaction.simulate {
		method = SIMULATE + "maintainContract"
	} else {
		method = CONTRACT + "maintainContract"
	}
	transaction.isMaintain = true
	param := transaction.Serialize()
	if transaction.simulate {
		return rpc.Call(method, param)
	}
	return rpc.CallByPolling(method, param, transaction.isPrivateTx)
}

// GetContractStatus ??????????????????
func (rpc *RPC) GetContractStatus(contractAddress string) (string, StdError) {
	method := CONTRACT + "getStatus"
	param := contractAddress
	data, err := rpc.call(method, param)
	if err != nil {
		return "", err
	}
	result := string([]byte(data))
	return result, nil
}

// GetContractStatusByName ??????????????????
func (rpc *RPC) GetContractStatusByName(contractName string) (string, StdError) {
	method := CONTRACT + "getStatusByCName"
	param := contractName
	data, err := rpc.call(method, param)
	if err != nil {
		return "", err
	}
	result := string([]byte(data))
	return result, nil
}

// GetCreator ?????????????????????
func (rpc *RPC) GetCreator(contractAddress string) (string, StdError) {
	method := CONTRACT + "getCreator"
	param := contractAddress
	data, err := rpc.call(method, param)
	if err != nil {
		return "", err
	}
	var accountAddress string
	if sysError := json.Unmarshal(data, &accountAddress); sysError != nil {
		return "", NewSystemError(err)
	}
	return accountAddress, nil
}

// GetCreatorByName ?????????????????????
func (rpc *RPC) GetCreatorByName(contractName string) (string, StdError) {
	method := CONTRACT + "getCreatorByCName"
	param := contractName
	data, err := rpc.call(method, param)
	if err != nil {
		return "", err
	}
	var accountAddress string
	if sysError := json.Unmarshal(data, &accountAddress); sysError != nil {
		return "", NewSystemError(err)
	}
	return accountAddress, nil
}

// GetCreateTime ????????????????????????
func (rpc *RPC) GetCreateTime(contractAddress string) (string, StdError) {
	method := CONTRACT + "getCreateTime"
	param := contractAddress
	data, err := rpc.call(method, param)
	if err != nil {
		return "", err
	}
	var dateTime string
	if sysError := json.Unmarshal(data, &dateTime); sysError != nil {
		return "", NewSystemError(err)
	}
	return dateTime, nil
}

// GetCreateTimeByName ????????????????????????
func (rpc *RPC) GetCreateTimeByName(contractName string) (string, StdError) {
	method := CONTRACT + "getCreateTimeByCName"
	param := contractName
	data, err := rpc.call(method, param)
	if err != nil {
		return "", err
	}
	var dateTime string
	if sysError := json.Unmarshal(data, &dateTime); sysError != nil {
		return "", NewSystemError(err)
	}
	return dateTime, nil
}

// GetDeployedList ??????????????????????????????
func (rpc *RPC) GetDeployedList(address string) ([]string, StdError) {
	method := CONTRACT + "getDeployedList"
	param := address
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}
	var result []string
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, NewSystemError(err)
	}
	return result, nil
}

// InvokeContractReturnHash for pressure test
// Deprecated:
func (rpc *RPC) InvokeContractReturnHash(transaction *Transaction) (string, StdError) {
	method := CONTRACT + "invokeContract"
	param := transaction.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return "", err
	}

	var hash string
	if sysErr := json.Unmarshal(data, &hash); sysErr != nil {
		return "", NewSystemError(sysErr)
	}

	return hash, nil
}

// SendTxReturnHash for pressure test
// Deprecated:
func (rpc *RPC) SendTxReturnHash(transaction *Transaction) (string, StdError) {
	method := TRANSACTION + "sendTransaction"
	param := transaction.Serialize()
	data, err := rpc.call(method, param)
	if err != nil {
		return "", err
	}

	var hash string
	if sysErr := json.Unmarshal(data, &hash); sysErr != nil {
		return "", NewSystemError(sysErr)
	}

	return hash, nil
}

// GetTransactionsByExtraID ??????extraID????????????
// extraId ????????????????????????????????????
func (rpc *RPC) GetTransactionsByExtraID(extraId []interface{}, txTo string, detail bool, mode int, metadata *Metadata) (*PageResult, StdError) {
	method := TRANSACTION + "getTransactionsByExtraID"
	filter := &Filter{ExtraId: extraId}
	if txTo != "" {
		filter.TxTo = txTo
	}
	param := newMapParam("filter", filter)
	param.addKV("detail", detail)
	param.addKV("mode", mode)
	if metadata != nil {
		param.addKV("metadata", metadata)
	}
	data, err := rpc.call(method, param.Serialize())
	if err != nil {
		return nil, err
	}

	var result PageResult
	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return &result, nil
}

// getTransactionsByFilter ??????Filter????????????
func (rpc *RPC) getTransactionsByFilter(filter *Filter, detail bool, mode int, metadata *Metadata) (*PageResult, StdError) {
	method := TRANSACTION + "getTransactionsByFilter"
	param := newMapParam("filter", filter)
	param.addKV("detail", detail)
	param.addKV("mode", mode)
	if metadata != nil {
		param.addKV("metadata", metadata)
	}

	data, err := rpc.call(method, param.Serialize())
	if err != nil {
		return nil, err
	}

	var result PageResult
	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return &result, nil
}

/*---------------------------------- sub ----------------------------------*/

// GetWebSocketClient ??????WebSocket?????????
func (rpc *RPC) GetWebSocketClient() *WebSocketClient {
	once.Do(func() {
		globalWebSocketClient = &WebSocketClient{
			conns:   make(map[int]*connectionWrapper, len(rpc.hrm.nodes)),
			hrm:     &rpc.hrm,
			rwMutex: sync.RWMutex{},
		}
	})

	return globalWebSocketClient
}

/*---------------------------------- mq ----------------------------------*/

// GetMqClient ??????mq?????????
func (rpc *RPC) GetMqClient() *MqClient {
	once.Do(func() {
		mqClient = &MqClient{
			mqConns: make(map[uint]*mqWrapper, len(rpc.hrm.nodes)),
			hrm:     &rpc.hrm,
		}
	})

	return mqClient
}

/*---------------------------------- archive ----------------------------------*/

// Snapshot makes the snapshot for given the future block number or current the latest block number.
// It returns the snapshot id for the client to query.
// blockHeight can use `latest`, means make snapshot now
func (rpc *RPC) Snapshot(blockHeight interface{}) (string, StdError) {
	method := ARCHIVE + "snapshot"

	data, stdErr := rpc.call(method, blockHeight)
	if stdErr != nil {
		return "", stdErr
	}

	var result string

	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return "", NewSystemError(sysErr)
	}

	return result, nil
}

// QuerySnapshotExist checks if the given snapshot existed, so you can confirm that
// the last step Archive.Snapshot is successful.
func (rpc *RPC) QuerySnapshotExist(filterID string) (bool, StdError) {
	method := ARCHIVE + "querySnapshotExist"

	data, stdErr := rpc.call(method, filterID)
	if stdErr != nil {
		return false, stdErr
	}

	var result bool

	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return false, NewSystemError(sysErr)
	}

	return result, nil
}

// CheckSnapshot will check that the snapshot is correct. If correct, returns true.
// Otherwise, returns false.
func (rpc *RPC) CheckSnapshot(filterID string) (bool, StdError) {
	method := ARCHIVE + "checkSnapshot"

	data, stdErr := rpc.call(method, filterID)
	if stdErr != nil {
		return false, stdErr
	}

	var result bool

	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return false, NewSystemError(sysErr)
	}

	return result, nil
}

// DeleteSnapshot delete snapshot by id
func (rpc *RPC) DeleteSnapshot(filterID string) (bool, StdError) {
	method := ARCHIVE + "deleteSnapshot"

	data, stdErr := rpc.call(method, filterID)
	if stdErr != nil {
		return false, stdErr
	}

	var result bool

	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return false, NewSystemError(sysErr)
	}

	return result, nil
}

// ListSnapshot returns all the existed snapshot information.
func (rpc *RPC) ListSnapshot() (Manifests, StdError) {
	method := ARCHIVE + "listSnapshot"

	data, stdErr := rpc.call(method)
	if stdErr != nil {
		return nil, stdErr
	}

	var result Manifests
	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	return result, nil
}

// ReadSnapshot returns the snapshot information for the given snapshot ID.
func (rpc *RPC) ReadSnapshot(filterID string) (*Manifest, StdError) {
	method := ARCHIVE + "readSnapshot"

	data, stdErr := rpc.call(method, filterID)
	if stdErr != nil {
		return nil, stdErr
	}

	var result Manifest
	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	return &result, nil
}

// Archive will archive data of the given snapshot. If successful, returns true.
func (rpc *RPC) Archive(filterID string, sync bool) (bool, StdError) {
	method := ARCHIVE + "archive"

	data, stdErr := rpc.call(method, filterID, sync)
	if stdErr != nil {
		return false, stdErr
	}

	var result bool

	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return false, NewSystemError(sysErr)
	}

	return result, nil
}

// ArchiveNoPredict used for archive to specific committed block-number
func (rpc *RPC) ArchiveNoPredict(filterID string) (string, StdError) {
	method := ARCHIVE + "archiveNoPredict"

	data, stdErr := rpc.call(method, filterID)
	if stdErr != nil {
		return "", stdErr
	}

	var result string

	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return "", NewSystemError(sysErr)
	}

	return result, nil
}

// Restore restores datas that have been archived for given snapshot. If successful, returns true.
func (rpc *RPC) Restore(filterID string, sync bool) (bool, StdError) {
	method := ARCHIVE + "restore"

	data, stdErr := rpc.call(method, filterID, sync)
	if stdErr != nil {
		return false, stdErr
	}

	var result bool
	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return false, NewSystemError(sysErr)
	}

	return result, nil
}

// RestoreAll restores all datas that have been archived. If successful, returns true.
func (rpc *RPC) RestoreAll(sync bool) (bool, StdError) {
	method := ARCHIVE + "restoreAll"

	data, stdErr := rpc.call(method, sync)
	if stdErr != nil {
		return false, stdErr
	}

	var result bool
	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return false, NewSystemError(sysErr)
	}

	return result, nil
}

// QueryArchiveExist checks if the given snapshot has been archived.
func (rpc *RPC) QueryArchiveExist(filterID string) (bool, StdError) {
	method := ARCHIVE + "queryArchiveExist"

	data, stdErr := rpc.call(method, filterID)
	if stdErr != nil {
		return false, stdErr
	}

	var result bool

	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return false, NewSystemError(sysErr)
	}

	return result, nil
}

// QueryArchive query archive status with the give snapshot.
func (rpc *RPC) QueryArchive(filterID string) (string, StdError) {
	method := ARCHIVE + "queryArchive"

	data, stdErr := rpc.call(method, filterID)
	if stdErr != nil {
		return "", stdErr
	}

	var result string

	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return "", NewSystemError(sysErr)
	}

	return result, nil
}

// Pending returns all pending snapshot requests in ascend sort.
func (rpc *RPC) Pending() ([]SnapshotEvent, StdError) {
	method := ARCHIVE + "pending"

	data, stdErr := rpc.call(method)
	if stdErr != nil {
		return nil, stdErr
	}

	var result []SnapshotEvent
	if sysErr := json.Unmarshal(data, &result); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}

	return result, nil
}

/*---------------------------------- cert ----------------------------------*/

// GetTCert ??????TCert
// Deprecated:
func (rpc *RPC) GetTCert(index uint) (string, StdError) {
	return rpc.hrm.getTCert(rpc.hrm.nodes[index].url)
}

/*---------------------------------- account ----------------------------------*/

// GetBalance ??????????????????
func (rpc *RPC) GetBalance(account string) (string, StdError) {
	account = chPrefix(account)
	method := ACCOUNT + "getBalance"
	param := account
	data, err := rpc.call(method, param)
	if err != nil {
		return "", err
	}

	var balance string
	if sysErr := json.Unmarshal(data, &balance); sysErr != nil {
		return "", NewSystemError(sysErr)
	}
	return balance, nil
}

// GetRoles ??????????????????
func (rpc *RPC) GetRoles(account string) ([]string, StdError) {
	account = chPrefix(account)
	method := ACCOUNT + "getRoles"
	param := account
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var roles []string
	if sysErr := json.Unmarshal(data, &roles); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return roles, nil
}

// GetAccountsByRole ????????????????????????
func (rpc *RPC) GetAccountsByRole(role string) ([]string, StdError) {
	method := ACCOUNT + "getAccountsByRole"
	param := role
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	var accounts []string
	if sysErr := json.Unmarshal(data, &accounts); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return accounts, nil
}

// GetContractStatus ??????????????????
func (rpc *RPC) GetAccountStatus(address string) (string, StdError) {
	method := ACCOUNT + "getStatus"
	param := address
	data, err := rpc.call(method, param)
	if err != nil {
		return "", err
	}
	result := string([]byte(data))
	return result, nil
}

/*---------------------------------- radar ----------------------------------*/

func (rpc *RPC) ListenContract(srcCode, addr string) (string, StdError) {
	method := RADAR + "registerContract"
	param := newMapParam("source", srcCode)
	param.addKV("addrsss", addr)

	data, err := rpc.call(method, param.Serialize())
	if err != nil {
		return "", err
	}

	return string(data), nil
}

/*---------------------------------- config ----------------------------------*/

func (rpc *RPC) GetProposal() (*ProposalRaw, StdError) {
	method := CONFIG + "getProposal"
	data, err := rpc.call(method)
	if err != nil {
		return nil, err
	}

	var proposal ProposalRaw
	if sysErr := json.Unmarshal(data, &proposal); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return &proposal, nil
}

func (rpc *RPC) GetConfig() (string, StdError) {
	method := CONFIG + "getConfig"
	data, err := rpc.call(method)
	if err != nil {
		return "", err
	}

	var config string
	if sysErr := json.Unmarshal(data, &config); sysErr != nil {
		return "", NewSystemError(sysErr)
	}
	return config, nil
}

func (rpc *RPC) GetHosts(role string) (map[string][]byte, StdError) {
	method := CONFIG + "getHosts"
	param := role
	data, err := rpc.call(method, param)
	if err != nil {
		return nil, err
	}

	hosts := make(map[string][]byte)
	if sysErr := json.Unmarshal(data, &hosts); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return hosts, nil
}

func (rpc *RPC) GetVSet() ([]string, StdError) {
	method := CONFIG + "getVSet"
	data, err := rpc.call(method)
	if err != nil {
		return nil, err
	}

	var vset []string
	if sysErr := json.Unmarshal(data, &vset); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return vset, nil
}

func (rpc *RPC) GetAllRoles() (map[string]int, StdError) {
	method := CONFIG + "getRoles"
	data, err := rpc.call(method)
	if err != nil {
		return nil, err
	}

	roles := make(map[string]int)
	if sysErr := json.Unmarshal(data, &roles); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return roles, nil
}

func (rpc *RPC) IsRoleExist(role string) (bool, StdError) {
	param := role
	method := CONFIG + "isRoleExist"
	data, err := rpc.call(method, param)
	if err != nil {
		return false, err
	}
	exist, er := strconv.ParseBool(string(data))
	if er != nil {
		return false, NewSystemError(er)
	}
	return exist, nil
}

// GetAddressByName get contract address by contract name
func (rpc *RPC) GetAddressByName(name string) (string, StdError) {
	param := name
	method := CONFIG + "getAddressByCName"
	data, err := rpc.call(method, param)
	if err != nil {
		return "", err
	}
	var addr string
	if sysErr := json.Unmarshal(data, &addr); sysErr != nil {
		return "", NewSystemError(sysErr)
	}
	return addr, nil
}

// GetNameByAddress get contract name by contract address
func (rpc *RPC) GetNameByAddress(address string) (string, StdError) {
	param := chPrefix(address)
	method := CONFIG + "getCNameByAddress"
	data, err := rpc.call(method, param)
	if err != nil {
		return "", err
	}
	var name string
	if sysErr := json.Unmarshal(data, &name); sysErr != nil {
		return "", NewSystemError(sysErr)
	}
	return name, nil
}

// GetAllCNS get all contract address to contract name mapping
func (rpc *RPC) GetAllCNS() (map[string]string, StdError) {
	method := CONFIG + "getAllCNS"
	data, err := rpc.call(method)
	if err != nil {
		return nil, err
	}
	all := make(map[string]string)
	if sysErr := json.Unmarshal(data, &all); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return all, nil
}

// SetAccount set account key for sign request
func (rpc *RPC) SetAccount(key account.Key) {
	rpc.im.key = key
}

// AddRoleForNode add roles for given address in node
func (rpc *RPC) AddRoleForNode(address string, roles ...string) StdError {
	method := AUTH + "addRole"
	_, err := rpc.call(method, address, roles)
	if err != nil {
		return err
	}
	return nil
}

// DeleteRoleFromNode delete roles from address in node
func (rpc *RPC) DeleteRoleFromNode(address string, roles ...string) StdError {
	method := AUTH + "deleteRole"
	_, err := rpc.call(method, address, roles)
	if err != nil {
		return err
	}
	return nil
}

// GetRoleFromNode get account role in node
func (rpc *RPC) GetRoleFromNode(address string) ([]string, StdError) {
	method := AUTH + "getRole"
	data, err := rpc.call(method, address)
	if err != nil {
		return nil, err
	}
	var roles []string
	if sysErr := json.Unmarshal(data, &roles); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return roles, nil
}

// GetAddressFromNode get address by role in node
func (rpc *RPC) GetAddressFromNode(role string) ([]string, StdError) {
	method := AUTH + "getAddress"
	data, err := rpc.call(method, role)
	if err != nil {
		return nil, err
	}
	var address []string
	if sysErr := json.Unmarshal(data, &address); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return address, nil
}

// GetAllRolesFromNode get address by role in node
func (rpc *RPC) GetAllRolesFromNode() ([]string, StdError) {
	method := AUTH + "getAllRoles"
	data, err := rpc.call(method)
	if err != nil {
		return nil, err
	}
	var address []string
	if sysErr := json.Unmarshal(data, &address); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return address, nil
}

// SetRulesInNode set inspector rules for auth api in node
func (rpc *RPC) SetRulesInNode(rules []*InspectorRule) StdError {
	method := AUTH + "setRules"
	_, err := rpc.call(method, rules)
	if err != nil {
		return err
	}
	return nil
}

// GetRulesFromNode get inspector rules for auth api in node
func (rpc *RPC) GetRulesFromNode() ([]*InspectorRule, StdError) {
	method := AUTH + "getRules"
	data, err := rpc.call(method)
	if err != nil {
		return nil, err
	}
	var rules []*InspectorRule
	if sysErr := json.Unmarshal(data, &rules); sysErr != nil {
		return nil, NewSystemError(sysErr)
	}
	return rules, nil
}
