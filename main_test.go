package main

import (
	"reflect"
	"testing"
)

var (
	dummyInput  = "vmess://eyJhZGQiOiJteS52Mi1yYXkuY29tIiwiYWlkIjoiMCIsImlkIjoiYWZlZjkyZDUtYjViOS00NmFhLWE5ODItMzkxNGZlNjJjMWM1IiwiaG9zdCI6Im1vYmllZHUudm4iLCJuZXQiOiJ3cyIsInBhdGgiOiIvZmFzdHNzaC8xMjA4LzYyZjViMTU4YTZjNTUvIiwicG9ydCI6IjQ0MyIsInBzIjoibXkudjItcmF5LmNvbSIsInRscyI6InRscyIsInR5cGUiOiJub25lIiwidiI6IjIifQ=="
	expectedCfg = &Config{
		VnextAddr:                 "my.v2-ray.com",
		VnextPort:                 443,
		VnextUserID:               "afef92d5-b5b9-46aa-a982-3914fe62c1c5",
		VnextUserAlterId:          0,
		StreamSettingNetwork:      "ws",
		StreamSettingSecurity:     "tls",
		StreamSettingTLSInsecure:  true,
		StreamSettingWSPath:       "/fastssh/1208/62f5b158a6c55/",
		StreamSettingWSHeaderHost: "mobiedu.vn",
	}
)

func TestParseConfig(t *testing.T) {
	cfg, err := ParseConfigV2ray(dummyInput)
	if err != nil {
		t.Error(err)
	}
	if want, got := expectedCfg, cfg; !reflect.DeepEqual(want, got) {
		t.Errorf("want config: %v, but got: %v", want, got)
	}
	t.Run("can handle invalid prefix", handleInvalidPrefix())
}
func handleInvalidPrefix() func(t *testing.T) {
	return func(t *testing.T) {
		invalidStr := "abcdec"
		_, err := ParseConfigV2ray(invalidStr)
		if want, got := ErrorInvalidFormatVMess, err; got != want {
			t.Errorf("Want error: %s, but got error: %s", want.Error(), got.Error())
		}
	}
}

// func handleGenerateTemplate() func(t *testing.T) {
// 	return func(t *testing.T) {
// 		flag.Set("")
// 	}
// }
