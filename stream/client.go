package stream

import (
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Downloader struct {
	client        *http.Client
	Url           string
	FilePrefix    string
	FileDirectory string

	logger *log.Entry
}

func (d Downloader) Download(downloadDuration time.Duration) error {
	d.logger.Infof("Download url=%s", d.Url)
	restDuration := downloadDuration
	var err error

	for retryNumber := 0; retryNumber < maxRetryNumber; retryNumber++ {
		d.logger = d.logger.WithField("retryNumber", retryNumber)
		restDuration, err = d.download(restDuration)
		if err == nil && restDuration <= 0 {
			return nil
		}
	}
	return errors.Wrap(err, "Fail downloading")
}

func (d Downloader) download(downloadDuration time.Duration) (time.Duration, error) {
	d.logger.Infof("Start downloading, durationSec=%f", downloadDuration.Seconds())

	response, err := d.makeRequest()
	if err != nil {
		return downloadDuration, err
	}
	defer response.Body.Close()

	streamFile, err := d.createFile()
	if err != nil {
		return downloadDuration, err
	}
	defer streamFile.Close()

	leftDuration, err := d.savingStream(downloadDuration, streamFile, response.Body)
	if err != nil || leftDuration != 0 {
		return leftDuration, err
	}

	d.logger.Info("Success downloaded")
	return 0, nil
}

func (d Downloader) makeRequest() (response *http.Response, err error) {
	response, err = d.client.Get(d.Url)
	if err != nil {
		d.logger.Errorf("Can not get response. err='%+v'", err)
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		d.logger.Errorf("Got incorrect response, status_code=%d", response.StatusCode)
		return nil, errors.New("incorrect response status code")
	}
	d.logger.Info("Got success response")
	return response, nil
}

func (d Downloader) createFile() (*os.File, error) {
	fileName := d.FilePrefix + time.Now().Format(DateFormat) + ".mp3"
	filePath := filepath.Join(d.FileDirectory, fileName)
	d.logger.Infof("Creating file=%s", filePath)

	file, err := os.Create(filePath)
	if err != nil {
		d.logger.Errorf("Fail in creating file, err='%+v'", err)
	}
	return file, err
}

func (d Downloader) savingStream(
	downloadDuration time.Duration, file io.Writer, responseBody io.Reader,
) (time.Duration, error) {
	startDownloadingTime := time.Now()
	for {
		duration := time.Now().Sub(startDownloadingTime)
		if duration > downloadDuration {
			break
		}

		if written, err := io.CopyN(file, responseBody, sizeKB); err != nil || written != sizeKB {
			leftDuration := downloadDuration - duration
			d.logger.Errorf("Fail in saving stream to file, written=%d leftDurationSec=%f err='%+v'",
				written, leftDuration.Seconds(), err,
			)
			return leftDuration, errors.New("fail copy stream")
		}
	}
	return 0, nil
}

func NewDownloader(url, filePrefix, directory string) (*Downloader, error) {
	logger := log.
		WithField("downloader_id", rand.Int())

	fDirectory, err := filepath.Abs(directory)
	if err != nil {
		logger.Errorf("Getting filepath err=%+v", err)
		return nil, err
	}

	// 1-execute, 2-write, 4-read (for owner, group, all)
	dirPerms := os.FileMode(666)  // rw-rw-rw-
	if err := os.MkdirAll(fDirectory, dirPerms); err != nil {
		logger.Errorf("Fail while created directory err=%+v", err)
		return nil, err
	}

	return &Downloader{
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   time.Second * 5,
				ResponseHeaderTimeout: time.Second * 5,
				ExpectContinueTimeout: time.Minute * 5,
				IdleConnTimeout:       time.Minute * 5, // keep-alive timeout
			},
		},
		Url:        url,
		FilePrefix: filePrefix,
		FileDirectory: fDirectory,

		logger: logger,
	}, nil
}
