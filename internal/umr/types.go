package umr

import "encoding/json"

// NetworkSource values observed in InfoDump.low.network_source. The router
// reports the currently active WAN here.
const (
	NetworkSourceWifiWAN = "wifiwan"
	NetworkSourceLTE     = "lte"
	NetworkSourceWAN     = "wan"
)

// session is the response payload of a successful session/login call.
type session struct {
	Token   string `json:"ubus_rpc_session"`
	Expires int    `json:"expires"`
	Timeout int    `json:"timeout"`
}

// InfoDump is the full dashboard payload (InfoDump). It is the union of the
// three Info*Dump levels plus client and gps. Prefer this over the standalone
// Info{Low,Medium,High}Dump calls: the composite returns a superset of fields.
type InfoDump struct {
	Low    InfoLow         `json:"low"`
	Medium InfoMedium      `json:"medium"`
	High   InfoHigh        `json:"high"`
	GPS    json.RawMessage `json:"gps"`
	Client InfoClient      `json:"client"`
}

// InfoLow holds firmware, SIM, and active-WAN identity fields. The InfoDump
// composite returns more fields here than the standalone InfoLowDump call,
// notably wifi_wan_* and geo_*.
type InfoLow struct {
	Firmware            string `json:"fw"`
	UptimeSeconds       int    `json:"uptime"`
	NetworkSource       string `json:"network_source"` // active WAN: wifiwan / lte / wan
	IP                  string `json:"ip"`
	GeoIP               string `json:"geo_ip"`
	GeoISP              string `json:"geo_isp"`
	WifiWANSSID         string `json:"wifi_wan_ssid"`
	WifiWANSignalLevel  int    `json:"wifi_wan_signal_level"`
	WifiWANStatusCode   int    `json:"wifi_wan_status_code"`
	OperatorName        string `json:"operator_name"`
	SignalLevel         int    `json:"signal_level"`
	ICCID               string `json:"iccid"`
	IMSI                string `json:"imsi"`
	SimInsertState      int    `json:"sim_insert_state"`
	LTEIPv4Address      string `json:"lte_ipv4_address"`
	APN                 string `json:"apn"`
	APNConfigurationSrc int    `json:"apn_configuration_source"`
	LTEAPNUsername      string `json:"lte_apn_username"`
	LTEAPNPassword      string `json:"lte_apn_password"`
	LTEAPNAuthType      string `json:"lte_apn_auth_type"`
	LTEPDPType          string `json:"lte_pdp_type"`
	LTERoamingAllowed   bool   `json:"lte_roaming_allowed"`
	LTEMode             string `json:"lte_mode"`
	LTEBand             string `json:"lte_band"`
	LTE4GBand           string `json:"lte_4g_band"`
}

// InfoMedium holds usage, throughput, sample, and resource fields.
type InfoMedium struct {
	DownloadUsage          int64 `json:"download_usage"`
	UploadUsage            int64 `json:"upload_usage"`
	TotalUsage             int64 `json:"total_usage"`
	UploadSpeed            int64 `json:"upload_speed"`   // bits per second
	DownloadSpeed          int64 `json:"download_speed"` // bits per second
	ClientNumbers          int   `json:"client_numbers"`
	WifiClients            int   `json:"wifi_clients"`
	Experience             int   `json:"experience"`
	PER                    int   `json:"per"`
	SampleTime             int64 `json:"sample_time"` // ms epoch
	CPU                    int   `json:"cpu"`         // percent
	Memory                 int   `json:"memory"`      // percent
	LatencyAvgMs           int   `json:"latency_avg_ms"`
	LatencyMaxMs           int   `json:"latency_max_ms"`
	LatencySampleCount     int   `json:"latency_sample_count"`
	LatencyPacketLossCount int   `json:"latency_packet_loss_count"`
	ResetUsageTimestamp    int64 `json:"reset_usage_timestamp"`
}

// InfoHigh is a superset of Low + Medium + extras (IPv6, ethernet ports,
// first client preview). The InfoDump.high view exposes everything visible
// on the portal's dashboard.
type InfoHigh struct {
	InfoLow
	InfoMedium

	LTEState  int    `json:"lte_state"`
	RSSI      int    `json:"rssi"`
	RSRQ      int    `json:"rsrq"`
	RSRP      int    `json:"rsrp"`
	RxChannel int    `json:"rx_channel"`
	TxChannel int    `json:"tx_channel"`
	Band      string `json:"band"`

	IPv6              []string       `json:"ipv6"`
	IPv6LAN           []string       `json:"ipv6_lan"`
	EthernetPorts     []EthernetPort `json:"ethernet_port"`
	Clients           int            `json:"clients"`
	SampleCount       int            `json:"sample_count"`
	SampleIntervalSec int            `json:"sample_interval_second"`
	DownloadUsageAvg  int64          `json:"download_usage_avg"`
	UploadUsageAvg    int64          `json:"upload_usage_avg"`
	Client0           *ClientEntry   `json:"client0,omitempty"`
}

// EthernetPort describes one physical port. state/speed/type are enums whose
// exact semantics aren't documented by Ubiquiti; values observed in practice:
// state 0 = unplugged, speed 3 = 1 Gbps, type 0 = LAN.
type EthernetPort struct {
	State int `json:"state"`
	Speed int `json:"speed"`
	Type  int `json:"type"`
}

// InfoClient is the InfoDump.client (or InfoClientDump.result) payload.
// Connected clients live as numbered keys (client0, client1, ...). Use List()
// to walk them in order.
type InfoClient struct {
	Clients           int           `json:"clients"`
	SampleCount       int           `json:"sample_count"`
	SampleIntervalSec int           `json:"sample_interval_second"`
	Extra             ClientEntries `json:"-"` // populated from clientN keys
}

// ClientEntries is the ordered list of client0..clientN.
type ClientEntries []ClientEntry

// ClientEntry describes one connected device (wifi or ethernet).
type ClientEntry struct {
	ID          int      `json:"id"`
	IP          string   `json:"ip"`
	IPv6        []string `json:"ipv6"`
	MAC         string   `json:"mac"`
	HostName    string   `json:"host_name"`
	Connection  string   `json:"connection"` // "ethernet" / "wifi"
	UptimeSec   int      `json:"uptime"`
	RxPackets   int64    `json:"rxPackets"`
	TxPackets   int64    `json:"txPackets"`
	RxBytes     int64    `json:"rxBytes"`
	TxBytes     int64    `json:"txBytes"`
	RxAggrBytes int64    `json:"rxAggrBytes"`
	TxAggrBytes int64    `json:"txAggrBytes"`
	TxRate      int      `json:"tx_rate"`
	RxRate      int      `json:"rx_rate"`
	LinkSpeed   int      `json:"link_speed"` // Mbps
}

// List returns the connected clients in stable client0..clientN order.
func (c *InfoClient) List() []ClientEntry {
	return c.Extra
}

// DeviceInfo is the GetDeviceInfo response.
type DeviceInfo struct {
	BoardRevision string   `json:"board_revision"`
	MAC           string   `json:"mac"`
	ModelName     string   `json:"model_name"`
	LTEModelName  string   `json:"lte_model_name"`
	CloudURL      string   `json:"cloud_url"`
	DeviceAC      string   `json:"device_ac"`
	BridgeMode    bool     `json:"bridge_mode"`
	WANIP         string   `json:"wan_ip"`
	LANIP         string   `json:"lan_ip"`
	LANIPv6       []string `json:"lan_ipv6"`
	WANIPv6       []string `json:"wan_ipv6"`
}

// DeviceStatus is the GetDeviceStatus response. When the WAN is wifi, the
// lte sub-object is mostly empty.
type DeviceStatus struct {
	InetSource InetSourceStatus `json:"inet_source"`
	Activated  bool             `json:"activated"`
	Online     bool             `json:"online"`
	SystemTime int64            `json:"system_time"`
	LTE        json.RawMessage  `json:"lte"`
}

// InetSourceStatus reports per-WAN connectivity. Each field is one of
// "active" or "disconnected" (other values may exist on future firmwares).
type InetSourceStatus struct {
	LTE  string `json:"lte"`
	LAN  string `json:"lan"`
	WiFi string `json:"wifi"`
}

// Status is the GetStatus response: it reports the device's cloud (MQTT)
// connectivity, not local health, despite the generic name.
type Status struct {
	Registered   bool   `json:"registered"`
	Error        int    `json:"error"`
	Connected    bool   `json:"connected"`
	MQTTEndpoint string `json:"mqtt_endpoint"`
	MQTTClientID string `json:"mqtt_client_id"`
}

// VpnStatus is the GetVpnStatus response. Type is one of "WIREGUARD",
// "OPENVPN", "IPSEC". Logs is gzip+base64 encoded.
type VpnStatus struct {
	Code           int    `json:"code"`
	Type           string `json:"type"`
	ServerIP       string `json:"server_ip"`
	Logs           string `json:"logs"`
	ClientTunnelIP string `json:"client_tunnel_ip"`
	Status         string `json:"status"` // e.g. "NOT_ESTABLISHED"
}

// SimProfile is the GetSimProfile response.
type SimProfile struct {
	ICCID   string          `json:"iccid"`
	Current json.RawMessage `json:"current"`
	Default json.RawMessage `json:"default"`
	OK      bool            `json:"ok"`
}

// Shadow names supported by ShadowReadLocal.
const (
	ShadowNameDeviceInfo   = "device_info"
	ShadowNameNetworks     = "networks"
	ShadowNamePowers       = "powers"
	ShadowNameBasicControl = "basic_control"
	ShadowNameVPN          = "vpn"
	ShadowNameFirewall     = "firewall"
)

// ShadowResponse is the envelope of every ShadowReadLocal call. Use the
// typed shadow methods on Client for known shadows.
type ShadowResponse struct {
	OK   bool            `json:"ok"`
	Data json.RawMessage `json:"data"`
}

// ShadowNetworks is the typed shape of the "networks" shadow.
type ShadowNetworks struct {
	BWLimitProfiles json.RawMessage      `json:"bw_limit_profiles"`
	DataUsage       json.RawMessage      `json:"data_usage"`
	ICCIDDataUsage  json.RawMessage      `json:"iccid_data_usage"`
	InternetSource  json.RawMessage      `json:"internet_source"`
	IPv6            json.RawMessage      `json:"ipv6"`
	LAN             json.RawMessage      `json:"lan"`
	LTE             json.RawMessage      `json:"lte"`
	NetworkSource   []NetworkSourceEntry `json:"network_source"`
	RouterMode      json.RawMessage      `json:"router_mode"`
	SpeedLimit      json.RawMessage      `json:"speed_limit"`
	WAN             json.RawMessage      `json:"wan"`
	WiFi            json.RawMessage      `json:"wifi"`
	WifiWAN         json.RawMessage      `json:"wifi_wan"`
}

// NetworkSourceEntry is one row in the configured WAN priority list.
type NetworkSourceEntry struct {
	Source string `json:"source"` // wifiwan / lte / wan
	Enable bool   `json:"enable"`
}

// ShadowDeviceInfoData is the typed shape of the "device_info" shadow.
type ShadowDeviceInfoData struct {
	Name        string `json:"name"`
	GPSDisabled bool   `json:"gps_disabled"`
}
