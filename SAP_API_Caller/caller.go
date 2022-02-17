package sap_api_caller

import (
	"fmt"
	"io/ioutil"
	"net/http"
	sap_api_output_formatter "sap-api-integrations-product-master-class-reads-rmq-kube/SAP_API_Output_Formatter"
	"strings"
	"sync"

	"github.com/latonaio/golang-logging-library-for-sap/logger"
	"golang.org/x/xerrors"
)

type RMQOutputter interface {
	Send(sendQueue string, payload map[string]interface{}) error
}

type SAPAPICaller struct {
	baseURL      string
	apiKey       string
	outputQueues []string
	outputter    RMQOutputter
	log          *logger.Logger
}

func NewSAPAPICaller(baseUrl string, outputQueueTo []string, outputter RMQOutputter, l *logger.Logger) *SAPAPICaller {
	return &SAPAPICaller{
		baseURL:      baseUrl,
		apiKey:       GetApiKey(),
		outputQueues: outputQueueTo,
		outputter:    outputter,
		log:          l,
	}
}

func (c *SAPAPICaller) AsyncGetProductMasterClass(product string, accepter []string) {
	wg := &sync.WaitGroup{}
	wg.Add(len(accepter))
	for _, fn := range accepter {
		switch fn {
		case "ProductGeneral":
			func() {
				c.ProductGeneral(product)
				wg.Done()
			}()
		default:
			wg.Done()
		}
	}

	wg.Wait()
}

func (c *SAPAPICaller) ProductGeneral(product string) {
	productGeneralData, err := c.callProductMasterClassSrvAPIRequirementProductGeneral("A_ClfnProduct", product)
	if err != nil {
		c.log.Error(err)
		return
	}
	err = c.outputter.Send(c.outputQueues[0], map[string]interface{}{"message": productGeneralData, "function": "ProductMasterClassProductGeneral"})
	if err != nil {
		c.log.Error(err)
		return
	}
	c.log.Info(productGeneralData)

	productClassData, err := c.callToProductClass(productGeneralData[0].ToProductClass)
	if err != nil {
		c.log.Error(err)
		return
	}
	err = c.outputter.Send(c.outputQueues[0], map[string]interface{}{"message": productClassData, "function": "ProductMasterClassProductClass"})
	if err != nil {
		c.log.Error(err)
		return
	}
	c.log.Info(productClassData)

	classDetailsData, err := c.callToClassDetails(productClassData[0].ToClassDetails)
	if err != nil {
		c.log.Error(err)
		return
	}
	err = c.outputter.Send(c.outputQueues[0], map[string]interface{}{"message": classDetailsData, "function": "ProductMasterClassClassDetails"})
	if err != nil {
		c.log.Error(err)
		return
	}
	c.log.Info(classDetailsData)

	productCharcData, err := c.callToProductCharc(productGeneralData[0].ToProductCharc)
	if err != nil {
		c.log.Error(err)
		return
	}
	err = c.outputter.Send(c.outputQueues[0], map[string]interface{}{"message": productCharcData, "function": "ProductMasterClassProductCharacteristic"})
	if err != nil {
		c.log.Error(err)
		return
	}
	c.log.Info(productCharcData)
}

func (c *SAPAPICaller) callProductMasterClassSrvAPIRequirementProductGeneral(api, product string) ([]sap_api_output_formatter.ProductGeneral, error) {
	url := strings.Join([]string{c.baseURL, "API_CLFN_PRODUCT_SRV", api}, "/")
	req, _ := http.NewRequest("GET", url, nil)

	c.setHeaderAPIKeyAccept(req)
	c.getQueryWithProductGeneral(req, product)

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return nil, xerrors.Errorf("API request error: %w", err)
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	data, err := sap_api_output_formatter.ConvertToProductGeneral(byteArray, c.log)
	if err != nil {
		return nil, xerrors.Errorf("convert error: %w", err)
	}
	return data, nil
}

func (c *SAPAPICaller) callToProductClass(url string) ([]sap_api_output_formatter.ToProductClass, error) {
	req, _ := http.NewRequest("GET", url, nil)
	c.setHeaderAPIKeyAccept(req)

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return nil, xerrors.Errorf("API request error: %w", err)
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	data, err := sap_api_output_formatter.ConvertToToProductClass(byteArray, c.log)
	if err != nil {
		return nil, xerrors.Errorf("convert error: %w", err)
	}
	return data, nil
}

func (c *SAPAPICaller) callToClassDetails(url string) (*sap_api_output_formatter.ToClassDetails, error) {
	req, _ := http.NewRequest("GET", url, nil)
	c.setHeaderAPIKeyAccept(req)

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return nil, xerrors.Errorf("API request error: %w", err)
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	data, err := sap_api_output_formatter.ConvertToToClassDetails(byteArray, c.log)
	if err != nil {
		return nil, xerrors.Errorf("convert error: %w", err)
	}
	return data, nil
}

func (c *SAPAPICaller) callToProductCharc(url string) ([]sap_api_output_formatter.ToProductCharc, error) {
	req, _ := http.NewRequest("GET", url, nil)
	c.setHeaderAPIKeyAccept(req)

	resp, err := new(http.Client).Do(req)
	if err != nil {
		return nil, xerrors.Errorf("API request error: %w", err)
	}
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	data, err := sap_api_output_formatter.ConvertToToProductCharc(byteArray, c.log)
	if err != nil {
		return nil, xerrors.Errorf("convert error: %w", err)
	}
	return data, nil
}

func (c *SAPAPICaller) setHeaderAPIKeyAccept(req *http.Request) {
	req.Header.Set("APIKey", c.apiKey)
	req.Header.Set("Accept", "application/json")
}

func (c *SAPAPICaller) getQueryWithProductGeneral(req *http.Request, product string) {
	params := req.URL.Query()
	params.Add("$filter", fmt.Sprintf("Product eq '%s'", product))
	req.URL.RawQuery = params.Encode()
}
