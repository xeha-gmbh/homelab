package bootstrap

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
)

func ParseImages(data map[string]interface{}) ([]*Image, error) {
	rawImages, isList := data[keyImages].([]interface{})
	if !isList {
		return nil, fmt.Errorf("expect key '%s' to be a list", keyImages)
	}

	images := make([]*Image, 0, len(rawImages))
	for _, oneRawImage := range rawImages {
		image := &Image{}
		if err := mapstructure.Decode(oneRawImage, image); err != nil {
			return nil, fmt.Errorf("failed to parse image: %s", err.Error())
		} else {
			images = append(images, image)
		}
	}

	return images, nil
}

type Image struct {
	Name 	string 	`yaml:"name"`
	Flavor 	string 	`yaml:"flavor"`
	Auto 	bool	`yaml:"auto"`
	UsbBoot	bool 	`yaml:"usb-boot"`
	Reuse 	bool 	`yaml:"reuse"`
}

const (
	keyImages = "images"
)