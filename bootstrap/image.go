package bootstrap

import (
	"fmt"
	"github.com/imulab/homelab/shared"
	"github.com/mitchellh/mapstructure"
)

func ParseImages(data map[string]interface{}) ([]*Image, error) {
	rawImages, isList := data[keyImages].([]interface{})
	if !isList {
		output.Fatal(shared.ErrParse.ExitCode,
			"Malformed config: {{index .error}}",
			map[string]interface{}{
				"event": "parse_error",
				"error": fmt.Sprintf("expect key %s to be a list.", keyImages),
				"key":   keyImages,
			})
		return nil, shared.ErrParse
	}

	images := make([]*Image, 0, len(rawImages))
	for _, oneRawImage := range rawImages {
		image := &Image{}
		if err := mapstructure.Decode(oneRawImage, image); err != nil {
			output.Fatal(shared.ErrParse.ExitCode,
				"Malformed config: failed to parse image. Cause: {{index .cause}}",
				map[string]interface{}{
					"event": "parse_error",
					"cause": err.Error(),
				})
			return nil, shared.ErrParse
		} else {
			images = append(images, image)
		}
	}

	return images, nil
}

type Image struct {
	Name    string `yaml:"name"`
	Flavor  string `yaml:"flavor"`
	Auto    bool   `yaml:"auto"`
	UsbBoot bool   `yaml:"usb-boot"`
	Reuse   bool   `yaml:"reuse"`
}

const (
	keyImages = "images"
)
