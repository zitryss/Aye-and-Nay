package compressor

import (
	"time"

	"github.com/spf13/viper"
)

func newImaginaryConfig() imaginaryConfig {
	return imaginaryConfig{
		host:    viper.GetString("compressor.imaginary.host"),
		port:    viper.GetString("compressor.imaginary.port"),
		times:   viper.GetInt("compressor.imaginary.retry.times"),
		pause:   viper.GetDuration("compressor.imaginary.retry.pause"),
		timeout: viper.GetDuration("compressor.imaginary.retry.timeout"),
	}
}

type imaginaryConfig struct {
	host    string
	port    string
	times   int
	pause   time.Duration
	timeout time.Duration
}

func newShortPixelConfig() shortPixelConfig {
	return shortPixelConfig{
		url:             viper.GetString("compressor.shortpixel.url"),
		url2:            viper.GetString("compressor.shortpixel.url2"),
		apiKey:          viper.GetString("compressor.shortpixel.apiKey"),
		times:           viper.GetInt("compressor.shortpixel.retry.times"),
		pause:           viper.GetDuration("compressor.shortpixel.retry.pause"),
		timeout:         viper.GetDuration("compressor.shortpixel.retry.timeout"),
		wait:            viper.GetString("compressor.shortpixel.wait"),
		uploadTimeout:   viper.GetDuration("compressor.shortpixel.uploadTimeout"),
		downloadTimeout: viper.GetDuration("compressor.shortpixel.downloadTimeout"),
		repeatIn:        viper.GetDuration("compressor.shortpixel.repeatIn"),
		restartIn:       viper.GetDuration("compressor.shortpixel.restartIn"),
	}
}

type shortPixelConfig struct {
	url             string
	url2            string
	apiKey          string
	times           int
	pause           time.Duration
	timeout         time.Duration
	wait            string
	uploadTimeout   time.Duration
	downloadTimeout time.Duration
	repeatIn        time.Duration
	restartIn       time.Duration
}
