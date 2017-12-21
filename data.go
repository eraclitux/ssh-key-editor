package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"strings"
	"time"
)

const (
	keyFilePath = "/.test_ssh/authorized_keys"
	gitHubURI   = "https://api.github.com/users"
)

// key models an entry in authorized_keys.
type key struct {
	ID, Cipher, PubKey, Date string
}

// tmplData populates html template.
type tmplData struct {
	Username, Hostname string
	Keys               []key
}

func parseLine(l string) key {
	var k key
	fields := strings.Fields(l)
	switch len(fields) {
	case 4:
		k = key{
			Cipher: fields[0],
			PubKey: fields[1],
			ID:     fields[2],
			Date:   fields[3],
		}
	case 3:
		// FIXME
	case 2:
		// FIXME
	}
	return k
}

func getKeys() ([]key, error) {
	var keys []key
	user, err := user.Current()
	if err != nil {
		return nil, err
	}
	// FIXME parametrize key file
	f, err := os.Open(user.HomeDir + keyFilePath)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		k := parseLine(scanner.Text())
		keys = append(keys, k)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return keys, nil
}

func assembleTmplData() (tmplData, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return tmplData{}, err
	}
	user, err := user.Current()
	if err != nil {
		return tmplData{}, err
	}
	keys, err := getKeys()
	if err != nil {
		return tmplData{}, err
	}
	data := tmplData{
		Username: user.Username,
		Hostname: hostname,
		Keys:     keys,
	}
	return data, nil
}

// extractKey sanitizes and parse data sended with textarea
// that may contain additional or missing things
// like another id at the end.
func extractKey(line string) (pubKey, cipher string, err error) {
	ee := strings.Fields(line)
	switch len(ee) {
	case 1:
		cipher = "unknown"
		pubKey = ee[0]
		return
	case 2, 3:
		cipher = ee[0]
		pubKey = ee[1]
		return
	}
	return "", "", errors.New("impossible to parse pubkey data")
}

func retriveFromGH(username string) (string, string, error) {
	resp, err := http.Get(gitHubURI + "/" + username + "/keys")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	data := []struct{ Key string }{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return "", "", err
	}
	if len(data) == 0 {
		return "", "", errors.New("pubkey not found on GitHub")
	}
	// Return only the first key.
	ll := strings.Fields(data[0].Key)
	return ll[1], ll[0], nil
}
func appendKey(keyID, cipher, pubKey string) error {
	user, err := user.Current()
	if err != nil {
		return err
	}
	// FIXME parametrize key file
	f, err := os.OpenFile(user.HomeDir+"/.test_ssh/authorized_keys", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	date := time.Now().UTC().Format(time.RFC3339)
	l := fmt.Sprintf("%s %s %s %s\n", cipher, pubKey, keyID, date)
	_, err = f.Write([]byte(l))
	if err != nil {
		return err
	}
	return nil
}
