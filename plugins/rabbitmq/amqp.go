package rabbitmq

import (
	"bytes"
	"encoding/json"
	log "k8s.io/klog/v2"
	"net/http"
	"strings"
)

type APIQueueInfo struct {
	Messsages int64 `json:"messages"`
}

func (r *RabbitMQPlugin) getQueueLengthFromAPI() (int64, error) {
	apiQueueInfo := APIQueueInfo{}
	err := r.doApiRequest(r.Url, &apiQueueInfo)
	if err != nil {
		log.Errorf("Error getting queue length from API: %s", err)
		return 0, err
	}
	return apiQueueInfo.Messsages, nil
}

func (r *RabbitMQPlugin) doApiRequest(uri string, apiQueueInfo *APIQueueInfo) error {
	req, err := r.buildRequest(uri)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	reader := new(bytes.Buffer)
	_, _ = reader.ReadFrom(resp.Body)
	return json.Unmarshal(reader.Bytes(), &apiQueueInfo)
}

func (r *RabbitMQPlugin) buildRequest(uri string) (*http.Request, error) {
	index := strings.LastIndex(uri, "/")
	vhost := uri[index:]
	uri = uri[:index]
	uri = uri + "/api/queues" + vhost + "/" + r.Queue
	return http.NewRequest("GET", uri, nil)
}
