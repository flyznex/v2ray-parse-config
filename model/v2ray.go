package model

type V2rayConfig struct {
	VnextAddr                 string
	VnextPort                 int
	VnextUserID               string
	VnextUserAlterId          int
	StreamSettingNetwork      string
	StreamSettingSecurity     string
	StreamSettingTLSInsecure  bool
	StreamSettingWSPath       string
	StreamSettingWSHeaderHost string
}
