package rabbitmq

import (
	"bytes"
	"encoding/json"
	"github.com/streadway/amqp"
	log "k8s.io/klog/v2"
	"net/http"
	"strings"
)

type APIQueueInfo struct {
	Messsages int64 `json:"messages"`
}

func getQueueLengthFromAPI(uri, name string) (int64, error) {
	apiQueueInfo := APIQueueInfo{}
	err := doApiRequest(uri, name, &apiQueueInfo)
	if err != nil {
		log.Errorf("Error getting queue length from API: %s", err)
		return 0, err
	}
	return apiQueueInfo.Messsages, nil
}

func doApiRequest(uri, name string, apiQueueInfo *APIQueueInfo) error {
	req, err := buildRequest(uri, name)
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	reader := new(bytes.Buffer)
	_, _ = reader.ReadFrom(resp.Body)
	return json.Unmarshal(reader.Bytes(), &apiQueueInfo)
}

func buildRequest(uri, name string) (*http.Request, error) {
	index := strings.LastIndex(uri, "/")
	vhost := uri[index:]
	uri = uri[:index]
	uri = uri + "/api/queues" + vhost + "/" + name
	return http.NewRequest("GET", uri, nil)
}

func getQueueLength(uri, name string) (int64, error) {
	if strings.HasPrefix(uri, "http") {
		return getQueueLengthFromAPI(uri, name)
	}
	conn, err := amqp.Dial(uri)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return 0, err
	}
	defer ch.Close()
	q, err := ch.QueueInspect(name)
	if err != nil {
		return 0, err
	}
	return int64(q.Messages), nil
}
