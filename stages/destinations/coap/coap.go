package coap

import (
	"context"
	"encoding/json"
	"github.com/dustin/go-coap"
	"github.com/streamsets/dataextractor/api"
	"github.com/streamsets/dataextractor/container/common"
	"github.com/streamsets/dataextractor/stages/stagelibrary"
	"log"
	"net/url"
)

const (
	LIBRARY            = "streamsets-datacollector-basic-lib"
	STAGE_NAME         = "com_streamsets_pipeline_stage_destination_coap_CoapClientDTarget"
	CONF_RESOURCE_URL  = "conf.resourceUrl"
	CONF_COAP_METHOD   = "conf.coapMethod"
	CONF_RESOURCE_TYPE = "conf.requestType"
	CONFIRMABLE        = "CONFIRMABLE"
	NONCONFIRMABLE     = "NONCONFIRMABLE"
	GET                = "GET"
	POST               = "POST"
	PUT                = "PUT"
	DELETE             = "DELETE"
)

type CoapClientDestination struct {
	resourceUrl string
	coapMethod  string
	requestType string
}

func init() {
	stagelibrary.SetCreator(LIBRARY, STAGE_NAME, func() api.Stage {
		return &CoapClientDestination{}
	})
}

func (c *CoapClientDestination) Init(ctx context.Context) error {
	stageContext := (ctx.Value("stageContext")).(common.StageContext)
	stageConfig := stageContext.StageConfig
	log.Println("[DEBUG] CoapClientDestination Init method")
	for _, config := range stageConfig.Configuration {
		if config.Name == CONF_RESOURCE_URL {
			c.resourceUrl = config.Value.(string)
		}

		if config.Name == CONF_COAP_METHOD {
			c.coapMethod = config.Value.(string)
		}

		if config.Name == CONF_RESOURCE_TYPE {
			c.requestType = config.Value.(string)
		}
	}

	return nil
}

func (c *CoapClientDestination) Write(batch api.Batch) error {
	log.Println("[DEBUG] CoapClientDestination Write method")
	for _, record := range batch.GetRecords() {
		err := c.sendRecordToSDC(record.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *CoapClientDestination) sendRecordToSDC(recordValue interface{}) error {
	jsonValue, err := json.Marshal(recordValue)
	if err != nil {
		return err
	}

	parsedURL, err := url.Parse(c.resourceUrl)
	if err != nil {
		return err
	}

	req := coap.Message{
		Type:    getCoapType(c.requestType),
		Code:    getCoapMethod(c.coapMethod),
		Payload: jsonValue,
	}
	req.SetPathString(parsedURL.Path)

	coapClient, err := coap.Dial("udp", parsedURL.Host)
	if err != nil {
		log.Printf("[ERROR] Error dialing: %v", err)
		return err
	}

	rv, err := coapClient.Send(req)
	if err != nil {
		log.Printf("[ERROR] Error sending request: %v", err)
		return err
	}

	if rv != nil {
		log.Printf("[DEBUG] Response payload: %s", rv.Payload)
	}

	return nil
}

func (h *CoapClientDestination) Destroy() error {
	return nil
}

func getCoapType(requestType string) coap.COAPType {
	switch requestType {
	case CONFIRMABLE:
		return coap.Confirmable
	case NONCONFIRMABLE:
		return coap.NonConfirmable
	}
	return coap.NonConfirmable
}

func getCoapMethod(coapMethod string) coap.COAPCode {
	switch coapMethod {
	case GET:
		return coap.GET
	case POST:
		return coap.POST
	case PUT:
		return coap.PUT
	case DELETE:
		return coap.DELETE
	}
	return coap.POST
}
