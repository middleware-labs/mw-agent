package frontend

import (
	"bytes"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"time"
)

type FrontendApi struct {
	Server string
	Token  string
	Logger *zap.Logger
}

func (f *FrontendApi) Request(method string, endpoint string, data map[string]any) (map[string]any, error) {
	url := f.Server + endpoint

	spaceClient := http.Client{
		Timeout: time.Second * 60, // Timeout after 2 seconds
	}

	jsonValue, _ := json.Marshal(data)
	//f.Logger.Info("requesting " + method + " " + url + " :: " + string(jsonValue))

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", f.Token)
	if err != nil {
		f.Logger.Fatal("http request create failed "+err.Error(), zap.Error(err))
	}

	res, getErr := spaceClient.Do(req)
	if getErr != nil {
		f.Logger.Fatal("http request do failed", zap.Error(getErr))
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	//f.Logger.Info("response body received " + string(body))
	if readErr != nil {
		return nil, errors.New("http request read failed:" + readErr.Error())
	}

	resp := map[string]any{}
	jsonErr := json.Unmarshal(body, &resp)
	if jsonErr != nil {
		return nil, errors.New("http request json parse failed: " + err.Error())
	}
	if _, ok := resp["success"].(bool); ok && resp["success"] == false {
		return nil, errors.New("http api response failed: " + resp["error"].(string))
	}
	if resp["data"] == nil {
		return nil, errors.New("data is nill")
	}
	return resp["data"].(map[string]any), nil
}
