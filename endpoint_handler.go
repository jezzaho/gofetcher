package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Structs storing Api response

type HistResponse struct {
	Configs []Config `json:"configs"`
	Items   []Item   `json:"items"`

	//helper map
	configIndexMap map[string]int
}

func (h *HistResponse) BuildIndex() {
	if h.configIndexMap != nil {
		return
	}
	h.configIndexMap = make(map[string]int, len(h.Configs))
	for i, c := range h.Configs {
		h.configIndexMap[c.ID] = i
	}

}
func (h *HistResponse) getConfigIndexByKPIType(name string) (int, bool) {
	for i, c := range h.Configs {
		if c.KPIType == name {
			return i, true
		}
	}
	return -1, false
}

func (h *HistResponse) GetValuesByConfigIndex(idx int) ([]FlexibleInt, error) {
	if idx <= 0 || idx >= len(h.Configs) {
		return nil, fmt.Errorf("config index %d out of bounds", idx)
	}
	values := make([]FlexibleInt, 0, len(h.Items))
	for _, item := range h.Items {
		if idx < len(item.Values) {
			values = append(values, item.Values[idx])
		} else {
			values = append(values, 0)
		}
	}
	return values, nil
}
func (h *HistResponse) GetValuesByKPIType(name string) ([]FlexibleInt, error) {
	idx, ok := h.getConfigIndexByKPIType(name)
	if !ok {
		return nil, fmt.Errorf("data source name %q not found", name)
	}
	return h.GetValuesByConfigIndex(idx)
}

type Config struct {
	ID              string `json:"id"`
	ProviderID      string `json:"providerId"`
	DataSourceType  string `json:"data_source_type"`
	DataSourceName  string `json:"data_source_name"`
	KPIType         string `json:"kpi_type"`
	DataType        string `json:"data_type"`
	Unit            string `json:"unit"`
	AggregationType string `json:"aggregation_type"`
}
type Item struct {
	Timestamp fmtTimeStamp  `json:"timestamp"`
	Values    []FlexibleInt `json:"values"`
}
type FlexibleInt int

func (fi *FlexibleInt) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*fi = 0
		return nil
	}
	var num int
	if err := json.Unmarshal(data, &num); err == nil {
		*fi = FlexibleInt(num)
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		parsed, err := strconv.Atoi(str)
		if err != nil {
			return err
		}
		*fi = FlexibleInt(parsed)
		return nil
	}
	return fmt.Errorf("invalid int value: %s", string(data))
}

type fmtTimeStamp struct {
	time.Time
}

func (ct *fmtTimeStamp) UnmarshalJSON(data []byte) error {
	// Try string (RFC3339)
	if data[0] == '"' {
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}
		t, err := time.Parse(time.RFC3339, str)
		if err != nil {
			return err
		}
		ct.Time = t
		return nil
	}

	// Try number (Unix timestamp)
	var rawNumber json.Number
	if err := json.Unmarshal(data, &rawNumber); err != nil {
		return err
	}

	// Try parsing as int64 (milliseconds)
	if ms, err := rawNumber.Int64(); err == nil {
		switch {
		case ms > 1e12: // likely milliseconds
			ct.Time = time.Unix(0, ms*int64(time.Millisecond))
		default: // likely seconds
			ct.Time = time.Unix(ms, 0)
		}
		return nil
	}

	return fmt.Errorf("invalid timestamp format: %s", string(data))
}

func doRequest[T any](method, endpoint string, headers map[string]string, queryParams map[string]string) (*T, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	q := u.Query()
	for k, v := range queryParams {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("request build error: %w", err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, body)
	}
	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	return &result, nil
}

func FetchHistData(params []string) (*HistResponse, error) {
	token, err := loadTokenFromFile(os.Getenv("TOKEN_CACHE"))
	if err != nil {
		return nil, err
	}

	return doRequest[HistResponse](
		"GET",
		os.Getenv("HIST_ENDPOINT"),
		map[string]string{
			"Authorization": "Bearer " + token.Token,
			//this might be broken
			"Xovis-Api-Version": "1",
			"Accept":            "application/json",
		},
		map[string]string{
			"from":             params[0],
			"to":               params[1],
			"granularity":      params[2],
			"tags":             params[3],
			"aggregation_type": params[4],
		},
	)
}
