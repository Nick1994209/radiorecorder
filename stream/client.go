package stream

import (
	"errors"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type Downloader struct {
	client     *http.Client
	Url        string
	FilePrefix string

	logger *log.Entry
}

func (d Downloader) Download(downloadDuration time.Duration) {
	d.download(downloadDuration, 0)
}

func (d Downloader) download(downloadDuration time.Duration, retryNumber int) {
	logger := d.logger.WithField("retry", retryNumber)

	if retryNumber >= maxRetryNumber {
		logger.Errorf("Number retries is maximum")
		return
	}

	logger.Info("Start downloading, durationSec=$f", downloadDuration.Seconds())

	response, err := d.makeRequest()
	if err != nil {
		d.download(downloadDuration, retryNumber+1)
		return
	}
	defer response.Body.Close()

	streamFile, err := d.createFile()
	if err != nil {
		return
	}
	defer streamFile.Close()

	leftDuration, err := d.savingStream(downloadDuration, streamFile, response.Body)
	if err != nil || leftDuration != 0 {
		d.download(leftDuration, retryNumber+1)
		return
	}

	logger.Info("Success downloaded")
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
	d.logger.Infof("Creating file=%s", fileName)
	file, err := os.Create(fileName)
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

func NewDownloader(url string, filePrefix string) Downloader {
	return Downloader{
		client:     &http.Client{Timeout: HTTPTimeout},
		Url:        url,
		FilePrefix: filePrefix,

		logger: log.WithField("url", url).WithField("downloader_id", rand.Int()),
	}
}
