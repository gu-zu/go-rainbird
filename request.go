package rainbird

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

type rbresponse struct {
	Id     int    `json:"id"`
	Result result `json:"result"`
}
type result struct {
	Length int    `json:"length"`
	Data   string `json:"data"`
}
type rbwifires struct {
	Id     int        `json:"id"`
	Result WifiResult `json:"result"`
}
type WifiResult struct {
	MacAddress     string `json:"macAddress"`
	Ip             string `json:"localIpAddress"`
	NetMask        string `json:"localNetmask"`
	Gateway        string `json:"localGateway"`
	Rssi           int    `json:"rssi"`
	SSID           string `json:"wifiSsid"`
	WifiPass       string `json:"wifiPassword"`
	WifiSec        string `json:"wifiSecurity"`
	ApTimeoutNoLan int    `json:"apTimeoutNoLan"`
	ApTimeoutIdle  int    `json:"apTimeoutIdle"`
	ApSec          string `json:"apSecurity"`
	StickVer       string `json:"stickVersion"`
}

// TODO - add paran for res length/type to check for error res
func (rb *Device) message(data string, rbres string) ([]byte, error) {
	body := packageMsg(rb.msgid, data)
	response, err := rb.send(body, 0)
	if err != nil {
		return nil, err
	}
	result := new(rbresponse)
	json.Unmarshal(response, result)
	if result.Id != rb.msgid-1 {
		log.Println("Incorrect reponse id", result.Id, rb.msgid)
		return nil, fmt.Errorf("incorrect response id: %d->%d", rb.msgid, result.Id)
	}
	dt := result.Result.Data
	if dt[:2] == "00" {
		reason, ok := map[string]string{"01": "doesn't exist(for this model)", "02": "incorrect param count"}[dt[4:6]]
		if !ok {
			reason = "unknown reason(" + dt[4:6] + ")"
		}
		return nil, fmt.Errorf("rainbird error response for command %s: %s", dt[2:4], reason)
	}
	if rbres != "" && dt[:2] != rbres {
		return nil, fmt.Errorf("rainbird unexpected response %s, expected %s", dt[:2], rbres)
	}
	output, err := hex.DecodeString(dt)
	return output, err //returning here anyway, so caller will check whether hex resulted in an error
}
func (rb *Device) methodmsg(data string) (*WifiResult, error) {
	body := packageMethod(rb.msgid, data)
	response, err := rb.send(body, 1)
	if err != nil {
		return nil, err
	}
	result := new(rbwifires)
	json.Unmarshal(response, result)
	if result.Id != rb.msgid-1 {
		log.Println("Incorrect reponse id", result.Id, rb.msgid)
		return nil, fmt.Errorf("incorrect response id: %d->%d", rb.msgid, result.Id)
	}
	return &result.Result, nil
}
func (rb *Device) send(input string, wf int) ([]byte, error) {
	body := bytes.NewReader(encrypt(input, rb.pass))
	rb.msgid++
	req, err := http.NewRequest("POST", "http://"+rb.ip+"/stick", body)
	if err != nil {
		log.Println("HTTP request make error:", err.Error())
		return nil, err
	}
	req.Header.Add("Content-Length", fmt.Sprint(body.Len()))
	req.Header.Add("Content-Type", "application/octet-stream")
	// "User-Agent": "RainBird/2.0 CFNetwork/811.5.4 Darwin/16.7.0",
	start := time.Now()
	res, err := http.DefaultClient.Do(req)
	if wf == 0 {
		log.Println("Request for", input[66:69], "took", time.Since(start)) // avg 1 second
	} else {
		log.Println("Request for wifiparms took", time.Since(start)) // avg 1 second
	}
	if err != nil {
		return nil, err
	}
	rbody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == 403 {
		return nil, fmt.Errorf("incorrect password %s %s", rb.ip, rb.pass)
	}
	if res.StatusCode != 200 {
		log.Println("HTTP error response", res.StatusCode, res.Status, rbody)
		return nil, fmt.Errorf("non 200 statusCode %d %s %s", res.StatusCode, res.Status, rbody)
	}
	return decrypt(rbody, rb.pass), nil
}

// only for tunnelSip commands
func packageMsg(id int, data string) string {
	// could be achieved with json.Marshal() but this method is easier and more efficient given the simple structure of the json message
	string := `{"id": ` + fmt.Sprint(id) + `, "method": "tunnelSip", "params": {"length": ` + fmt.Sprint(len(data)/2) + `, "data": "` + data + `"}, "jsonrpc": "2.0"}`
	return string
}
func packageMethod(id int, data string) string {
	string := `{"id": ` + fmt.Sprint(id) + `, "method": "` + data + `", "params": {}, "jsonrpc": "2.0"}`
	return string
}

func encrypt(data string, pass string) []byte {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	paddedData := padData(data + "\x00\x10")
	block, err := aes.NewCipher(sha256hash(pass))
	if err != nil {
		log.Fatal("failed to initialize cipher", err)
	}
	cipher := cipher.NewCBCEncrypter(block, randBytes)
	result := make([]byte, len(paddedData))
	cipher.CryptBlocks(result, []byte(paddedData))
	return bytes.Join([][]byte{sha256hash(data), randBytes, result}, nil)
}

func decrypt(data []byte, pass string) []byte {
	block, err := aes.NewCipher(sha256hash(pass))
	if err != nil {
		log.Fatal("failed to initialize cipher", err)
	}
	cipher := cipher.NewCBCDecrypter(block, data[32:48])
	result := make([]byte, len(data)-48)
	cipher.CryptBlocks(result, data[48:])
	return bytes.Trim(result, "\x10\x0A\x00")
}

func padData(data string) string {
	BLOCK_SIZE := 16
	charsToAdd := BLOCK_SIZE - (len(data) % BLOCK_SIZE)
	return data + strings.Repeat("\x10", charsToAdd)
}

func sha256hash(src string) []byte {
	hash := sha256.New()
	hash.Write([]byte(src))
	return hash.Sum(nil)
}
