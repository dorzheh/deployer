package bundle

import (
	"errors"

	"github.com/dorzheh/deployer/config/common/xmlinput"
	"github.com/dorzheh/deployer/deployer"
	"github.com/dorzheh/deployer/utils"
)

type BundleStrategy interface {
	Parse(*deployer.CommonData, deployer.HostinfoDriver, *xmlinput.XMLInputData) (map[string]interface{}, error)
}

type Parser struct {
	strategy BundleStrategy
	data     interface{}
}

func NewParser(bundleConfigFile string, s BundleStrategy) (*Parser, error) {
	if s == nil {
		return nil, utils.FormatError(errors.New("bundle strategy is nil"))
	}

	data, err := utils.ParseXMLFile(bundleConfigFile, s)
	if err != nil {
		return nil, utils.FormatError(err)
	}
	return &Parser{s, data}, nil
}

func (p *Parser) Parse(d *deployer.CommonData, hidriver deployer.HostinfoDriver, xid *xmlinput.XMLInputData) (map[string]interface{}, error) {
	return p.strategy.Parse(d, hidriver, xid)
}

func NewParserBuff(bundleConfigStream []byte, s BundleStrategy) (*Parser, error) {
	if s == nil {
		return nil, utils.FormatError(errors.New("bundle strategy is nil"))
	}
	data, err := utils.ParseXMLBuff(bundleConfigStream, s)
	if err != nil {
		return nil, utils.FormatError(err)
	}
	return &Parser{s, data}, nil
}
