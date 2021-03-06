package pcopy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadRawConfig_WithCommentSuccess(t *testing.T) {
	config, err := loadRawConfig(strings.NewReader(`WebUI true
# WebUI false`))
	if err != nil {
		t.Fatal(err)
	}
	if config["WebUI"] != "true" {
		t.Fatalf("expected %s, got %s", "true", config["WebUI"])
	}
}

func TestLoadRawConfig_OverrideSuccess(t *testing.T) {
	config, err := loadRawConfig(strings.NewReader(`WebUI true
WebUI false`))
	if err != nil {
		t.Fatal(err)
	}
	if config["WebUI"] != "false" {
		t.Fatalf("expected %s, got %s", "false", config["WebUI"])
	}
}

func TestLoadRawConfig_TrimTrailingSpaceSuccess(t *testing.T) {
	config, err := loadRawConfig(strings.NewReader(`WebUI "true"    `))
	if err != nil {
		t.Fatal(err)
	}
	if config["WebUI"] != `"true"` {
		t.Fatalf("expected %s, got %s", "", config["WebUI"])
	}
}

func TestLoadRawConfig_EmptyValue1Success(t *testing.T) {
	config, err := loadRawConfig(strings.NewReader(`WebUI`))
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := config["WebUI"]; !ok || v != "" {
		t.Fatalf("expected %s, got %s (ok: %t)", "", config["WebUI"], ok)
	}
}

func TestLoadRawConfig_EmptyValue2Success(t *testing.T) {
	config, err := loadRawConfig(strings.NewReader(`WebUI   `)) // Trailing spaces on empty value
	if err != nil {
		t.Fatal(err)
	}
	if v, ok := config["WebUI"]; !ok || v != "" {
		t.Fatalf("expected %s, got %s (ok: %t)", "", config["WebUI"], ok)
	}
}

func TestLoadConfig_EmptyFileSuccess(t *testing.T) {
	config, err := loadConfig(strings.NewReader(``))
	if err != nil {
		t.Fatal(err)
	}
	if config.ClipboardDir != DefaultClipboardDir {
		t.Fatalf("expected %s, got %s", DefaultClipboardDir, config.ClipboardDir)
	}
	if !config.WebUI {
		t.Fatalf("expected %t, got %t", true, config.WebUI)
	}
}

func TestLoadConfig_AllTheThingsSuccess(t *testing.T) {
	dir := t.TempDir()

	// This is a test key, don't freak out!
	pemKey := `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIHf6DNxfzPdOtM3vw/wW1hTaShp/Z1t0eT7RRak/S39doAoGCCqGSM49
AwEHoUQDQgAEp60xIGJbAUAmUe+KP9KB8ge4B+vJTKnMSctysQnG+fKOCTc9q7EX
xmNBMaTK3zXTdMev+TiCfmljflB7ZTkjTw==
-----END EC PRIVATE KEY-----`
	keyFile := filepath.Join(dir, "key.key")
	ioutil.WriteFile(keyFile, []byte(pemKey), 0700)

	pemCert := `-----BEGIN CERTIFICATE-----
MIIBMjCB2KADAgECAhAmIv+vEcI8iwP/TR4G3MavMAoGCCqGSM49BAMCMBAxDjAM
BgNVBAMTBXBjb3B5MB4XDTIwMTIyMTE2MDE1NVoXDTIzMTIyODE2MDE1NVowEDEO
MAwGA1UEAxMFcGNvcHkwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNCAASnrTEgYlsB
QCZR74o/0oHyB7gH68lMqcxJy3KxCcb58o4JNz2rsRfGY0ExpMrfNdN0x6/5OIJ+
aWN+UHtlOSNPoxQwEjAQBgNVHREECTAHggVwY29weTAKBggqhkjOPQQDAgNJADBG
AiEA1W0sKuPLyxoW0QTn0jovq9cAzT4IT5HaGeX8Z5rWlE4CIQCGn1yMReAETlWB
D1OY3Axih+rz7mF2xHK20TxRuy1sqw==
-----END CERTIFICATE-----`
	certFile := filepath.Join(dir, "cert.crt")
	ioutil.WriteFile(certFile, []byte(pemCert), 0700)

	config, err := loadConfig(strings.NewReader(fmt.Sprintf(`
ListenAddr :1234
ServerAddr hi.com
Key Osz6osE1fRRirA==:XEBZJjB/7w4eCugzQSkwGMe8QW4nbsPvPMlle1wvW4I=
KeyFile %s
CertFile %s
ClipboardDir %s
ClipboardSizeLimit 10M
ClipboardCountLimit 101
FileSizeLimit 123k
FileExpireAfter 10d
WebUI false
`, keyFile, certFile, dir)))
	if err != nil {
		t.Fatal(err)
	}
	assertStrEquals(t, ":1234", config.ListenAddr)
	assertStrEquals(t, "hi.com:2586", config.ServerAddr)
	assertBytesEquals(t, fromBase64(t, "Osz6osE1fRRirA=="), config.Key.Salt)
	assertBytesEquals(t, fromBase64(t, "XEBZJjB/7w4eCugzQSkwGMe8QW4nbsPvPMlle1wvW4I="), config.Key.Bytes)
	assertStrEquals(t, keyFile, config.KeyFile)
	assertStrEquals(t, certFile, config.CertFile)
	assertStrEquals(t, dir, config.ClipboardDir)
	assertInt64Equals(t, 10*1024*1024, config.ClipboardSizeLimit)
	assertInt64Equals(t, 101, int64(config.ClipboardCountLimit))
	assertInt64Equals(t, 123*1024, config.FileSizeLimit)
	assertInt64Equals(t, 10*24, int64(config.FileExpireAfter.Hours()))
	assertBoolEquals(t, false, config.WebUI)
}

func TestParseDuration_ZeroSuccess(t *testing.T) {
	d, err := parseDuration("0")
	if err != nil {
		t.Fatal(err)
	}
	if d != 0 {
		t.Fatalf("expected %d, got %d", 0, d)
	}
}

func TestParseDuration_WithDaysSuccess(t *testing.T) {
	d, err := parseDuration("10d")
	if err != nil {
		t.Fatal(err)
	}
	if d != 10*24*time.Hour {
		t.Fatalf("expected %d, got %d", 10*24*time.Hour, d)
	}
}

func TestParseDuration_WithoutDaysSuccess(t *testing.T) {
	d, err := parseDuration("10h5m")
	if err != nil {
		t.Fatal(err)
	}
	if d != 10*time.Hour+5*time.Minute {
		t.Fatalf("expected %d, got %d", 10*time.Hour+5*time.Minute, d)
	}
}

func TestParseDuration_WithDaysAndHoursFailure(t *testing.T) {
	_, err := parseDuration("10d1h") // not supported
	if err == nil {
		t.Fatalf("expected error, got none")
	}
}

func TestExpandServerAddr_Expand(t *testing.T) {
	actual := ExpandServerAddr("myhost")
	expected := "myhost:2586"
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}

func TestExpandServerAddr_NoExpand(t *testing.T) {
	actual := ExpandServerAddr("myhost:1234")
	expected := "myhost:1234"
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}

func TestCollapseServerAddr_Collapse(t *testing.T) {
	actual := CollapseServerAddr("myhost:2586")
	expected := "myhost"
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}

func TestCollapseServerAddr_NoCollapse(t *testing.T) {
	actual := CollapseServerAddr("myhost:1234")
	expected := "myhost:1234"
	if actual != expected {
		t.Fatalf("expected %s, got %s", expected, actual)
	}
}

func TestConfig_GenerateURLUnprotected(t *testing.T) {
	config := NewConfig()
	config.ServerAddr = "some-host.com"

	url, err := config.GenerateURL("/some-path", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	assertStrEquals(t, "https://some-host.com:2586/some-path", url)
}

func TestConfig_GenerateURLProtected(t *testing.T) {
	config := NewConfig()
	config.ServerAddr = "some-host.com"
	config.Key = &Key{Salt: []byte("some salt"), Bytes: []byte("16 bytes exactly")}

	url, err := config.GenerateURL("/some-path", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(url, "https://some-host.com:2586/some-path?a=SE1BQyA") {
		t.Fatalf("expected URL mismatched, got %s", url)
	}
	// TODO This should actually validate the HMAC, but the authorize() method is in server.go
}

func TestConfig_GenerateClipURLUnprotected(t *testing.T) {
	config := NewConfig()
	config.ServerAddr = "some-host.com"

	url, err := config.GenerateClipURL("some-id", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	assertStrEquals(t, "https://some-host.com:2586/some-id", url)
}

func TestConfig_WriteFileAllTheThings(t *testing.T) {
	config := NewConfig()
	config.ServerAddr = "some-host.com"
	config.ListenAddr = ":8888"
	config.Key = &Key{Salt: []byte("some salt"), Bytes: []byte("16 bytes exactly")}
	config.CertFile = "some cert file"
	config.KeyFile = "some key file"
	config.ClipboardDir = "/tmp/clipboarddir"
	config.ClipboardCountLimit = 1234
	config.ClipboardSizeLimit = 9876
	config.FileSizeLimit = 777
	config.FileExpireAfter = time.Hour
	config.WebUI = false

	filename := filepath.Join(t.TempDir(), "some.conf")
	if err := config.WriteFile(filename); err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	contents := string(b)
	assertStrContains(t, contents, "ServerAddr some-host.com")
	assertStrContains(t, contents, "ListenAddr :8888")
	assertStrContains(t, contents, "Key c29tZSBzYWx0:MTYgYnl0ZXMgZXhhY3RseQ==")
	assertStrContains(t, contents, "CertFile some cert file")
	assertStrContains(t, contents, "KeyFile some key file")
	assertStrContains(t, contents, "ClipboardDir /tmp/clipboarddir")
	assertStrContains(t, contents, "ClipboardCountLimit 1234")
	assertStrContains(t, contents, "ClipboardSizeLimit 9876")
	assertStrContains(t, contents, "FileSizeLimit 777")
	assertStrContains(t, contents, "FileExpireAfter 1h")
	assertStrContains(t, contents, "WebUI false")
}

func TestConfig_WriteFileNoneOfTheThings(t *testing.T) {
	config := NewConfig()

	filename := filepath.Join(t.TempDir(), "some.conf")
	if err := config.WriteFile(filename); err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	contents := string(b)
	assertStrContains(t, contents, "# ServerAddr")
	assertStrContains(t, contents, "ListenAddr :2586")
	assertStrContains(t, contents, "# Key")
	assertStrContains(t, contents, "# CertFile")
	assertStrContains(t, contents, "# KeyFile")
	assertStrContains(t, contents, "ClipboardDir /var/cache/pcopy")
	assertStrContains(t, contents, "# ClipboardCountLimit")
	assertStrContains(t, contents, "# ClipboardSizeLimit")
	assertStrContains(t, contents, "# FileSizeLimit")
	assertStrContains(t, contents, "FileExpireAfter 7d")
	assertStrContains(t, contents, "# WebUI")
}

func TestConfig_LoadConfigFromFileFailedDueToMissingCert(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "some.conf")
	contents := `ListenAddr :1234
CertFile some.crt
`
	if err := ioutil.WriteFile(filename, []byte(contents), 0700); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfigFromFile(filename)
	if err == nil {
		t.Fatalf("expected error due to missing cert, got none")
	}
}

func TestConfigStore_FileFromName(t *testing.T) {
	dir := t.TempDir()
	store := newConfigStoreWithDir(dir)
	file := store.FileFromName("work")
	assertStrEquals(t, dir+"/work.conf", file)
}

func TestConfigStore_All(t *testing.T) {
	dir := t.TempDir()
	f1, _ := os.Create(dir + "/work.conf")
	f1.Close()
	f2, _ := os.Create(dir + "/default.conf")
	f2.Close()
	store := newConfigStoreWithDir(dir)
	configs := store.All()
	if len(configs) != 2 {
		t.Fatalf("expected two configs, got %d", len(configs))
	}
	_, ok1 := configs[dir+"/work.conf"]
	if !ok1 {
		t.Fatalf("expected 'work' entry, but didn't have one")
	}
	_, ok2 := configs[dir+"/default.conf"]
	if !ok2 {
		t.Fatalf("expected 'default' entry, but didn't have one")
	}
}

func TestParseSize_10GSuccess(t *testing.T) {
	s, err := parseSize("10G")
	if err != nil {
		t.Fatal(err)
	}
	assertInt64Equals(t, 10*1024*1024*1024, s)
}

func TestParseSize_10MUpperCaseSuccess(t *testing.T) {
	s, err := parseSize("10M")
	if err != nil {
		t.Fatal(err)
	}
	assertInt64Equals(t, 10*1024*1024, s)
}

func TestParseSize_10kLowerCaseSuccess(t *testing.T) {
	s, err := parseSize("10k")
	if err != nil {
		t.Fatal(err)
	}
	assertInt64Equals(t, 10*1024, s)
}

func TestParseSize_FailureInvalid(t *testing.T) {
	_, err := parseSize("not a size")
	if err == nil {
		t.Fatalf("expected error, but got none")
	}
}

func TestExtractClipboard(t *testing.T) {
	assertStrEquals(t, "myclip", ExtractClipboard("/etc/pcopy/myclip.conf"))
}

func TestDefaultCertFile_MustNotExist(t *testing.T) {
	assertStrEquals(t, "/etc/pcopy/myclip.crt", DefaultCertFile("/etc/pcopy/myclip.conf", false))
}

func TestDefaultCertFile_MustExistSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "myclip.conf")
	expectedCertFile := filepath.Join(tmpDir, "myclip.crt")
	if err := ioutil.WriteFile(expectedCertFile, []byte("something"), 0700); err != nil {
		t.Fatal(err)
	}
	assertStrEquals(t, expectedCertFile, DefaultCertFile(configFile, true))
}

func TestDefaultKeyFile_MustNotExist(t *testing.T) {
	assertStrEquals(t, "/etc/pcopy/myclip.key", DefaultKeyFile("/etc/pcopy/myclip.conf", false))
}
