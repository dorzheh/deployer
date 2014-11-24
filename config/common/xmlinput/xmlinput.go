package xmlinput

import (
	"errors"

	"github.com/dorzheh/deployer/utils"
)

func ParseXMLInput(config string) (*XMLInputData, error) {
	d, err := utils.ParseXMLFile(config, new(XMLInputData))
	if err != nil {
		return nil, err
	}
	if err := verify(d.(*XMLInputData)); err != nil {
		return nil, err
	}
	return d.(*XMLInputData), nil
}

func ParseXMLInputMock(data []byte) (*XMLInputData, error) {
	d, err := utils.ParseXMLBuff(data, new(XMLInputData))
	if err != nil {
		return nil, err
	}
	if err := verify(d.(*XMLInputData)); err != nil {
		return nil, err
	}
	return d.(*XMLInputData), nil
}

func verify(data *XMLInputData) error {
	if data.Nets.Configure {
		for _, net := range data.Networks {
			seenDirect := false
			seenPassthrough := false
			for _, mode := range net.Modes {
				switch mode.Type {
				case ConTypeDirect:
					seenDirect = true
				case ConTypePassthrough:
					seenPassthrough = true
				case ConTypeBridged:
				default:
					return errors.New("unexpected mode " + string(mode.Type))
				}
			}
			if seenDirect && seenPassthrough {
				return errors.New("either \"direct\" or \"passthrough\" permitted")
			}
		}
	}
	return nil
}
