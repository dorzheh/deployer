package xmlinput

import (
	"errors"

	"github.com/dorzheh/deployer/utils"
)

// ParseXMLInputFile is responsible for parsing a given configuration file
// representing appropriate input data to a srtuctured form
func ParseXMLInput(config string) (*XMLInputData, error) {
	d, err := utils.ParseXMLFile(config, new(XMLInputData))
	if err != nil {
		return nil, utils.FormatError(err)
	}
	if err := verify(d.(*XMLInputData)); err != nil {
		return nil, utils.FormatError(err)
	}
	return d.(*XMLInputData), nil
}

// ParseXMLInputBuf is responsible for parsing a given stream of bytes
// representing appropriate input data to a srtuctured form
func ParseXMLInputBuf(data []byte) (*XMLInputData, error) {
	d, err := utils.ParseXMLBuff(data, new(XMLInputData))
	if err != nil {
		return nil, utils.FormatError(err)
	}
	if err := verify(d.(*XMLInputData)); err != nil {
		return nil, utils.FormatError(err)
	}
	return d.(*XMLInputData), nil
}

func verify(data *XMLInputData) error {
	if data.Networks.Configure {
		for _, net := range data.Networks.Configs {
			seenDirect := false
			seenPassthrough := false
			for _, mode := range net.Modes {
				switch mode.Type {
				case ConTypeDirect:
					seenDirect = true
				case ConTypePassthrough:
					seenPassthrough = true
				case ConTypeBridged, ConTypeOVS, ConTypeVirtualNetwork, ConTypeSRIOV:
				default:
					return utils.FormatError(errors.New("unexpected mode " + string(mode.Type)))
				}
			}
			if seenDirect && seenPassthrough {
				return utils.FormatError(errors.New("either \"direct\" or\"passthrough\" permitted"))
			}
		}
		for _, nic := range data.HostNics.Allowed {
			if nic.Disjunction && nic.Priority {
				return utils.FormatError(errors.New("either \"priority\" or\"disjunction\" permitted"))
			}
		}
	}
	return nil
}
