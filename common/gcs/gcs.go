package gcs

import (
	"bytes"
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"io"
	"time"
)

type ServiceAccountKeyJSON struct {
	Type           string `json:"type"`
	ProjectID      string `json:"project_id"`
	PrivateKeyID   string `json:"private_key_id"`
	PrivateKey     string `json:"private_key"`
	ClientEmail    string `json:"client_email"`
	ClientID       string `json:"client_id"`
	AuthURI        string `json:"auth_uri"`
	TokenURI       string `json:"token_uri"`
	AuthProvider   string `json:"auth_provider"`
	ClientSecret   string `json:"client_secret"`
	UniverseDomain string `json:"universe_domain"`
}

type GCSClient struct {
	ServiceAccountKeyJSON ServiceAccountKeyJSON
	BucketName            string
}

type IGCSlient interface {
	UploadFile(context.Context, string, []byte) (string, error)
}

func NewGCSClient(serviceAccountKeyJSON ServiceAccountKeyJSON, bucketName string) IGCSlient {
	return &GCSClient{
		ServiceAccountKeyJSON: serviceAccountKeyJSON,
		BucketName:            bucketName,
	}
}

func (g *GCSClient) CreateClient(ctx context.Context) (*storage.Client, error) {
	reqBodyBytes := new(bytes.Buffer)
	err := json.NewEncoder(reqBodyBytes).Encode(g.ServiceAccountKeyJSON)
	if err != nil {
		logrus.Errorf("failed to encode service account key: %v", err)
		return nil, err
	}

	jsonBytes := reqBodyBytes.Bytes()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(jsonBytes))
	if err != nil {
		logrus.Errorf("failed to create storage client: %v", err)
		return nil, err
	}

	return client, nil
}

func (g *GCSClient) UploadFile(ctx context.Context, fileName string, data []byte) (string, error) {
	var (
		contentType      = "application/octet-stream"
		timeoutInSeconds = 60
	)

	client, err := g.CreateClient(ctx)
	if err != nil {
		logrus.Errorf("failed to create storage client: %v", err)
		return "", err
	}

	defer func(client *storage.Client) {
		err := client.Close()
		if err != nil {
			logrus.Errorf("failed to close storage client: %v", err)
			return
		}
	}(client)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutInSeconds)*time.Second)
	defer cancel()

	bucket := client.Bucket(g.BucketName)
	obj := bucket.Object(fileName)
	buffer := bytes.NewBuffer(data)

	writter := obj.NewWriter(ctx)
	writter.ChunkSize = 0

	_, err = io.Copy(writter, buffer)
	if err != nil {
		logrus.Errorf("failed to copy data to object: %v", err)
		return "", err
	}

	err = writter.Close()
	if err != nil {
		logrus.Errorf("failed to close object: %v", err)
		return "", err
	}

	_, err = obj.Update(ctx, storage.ObjectAttrsToUpdate{ContentType: contentType})
	if err != nil {
		logrus.Errorf("failed to update object: %v", err)
		return "", err
	}

	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", g.BucketName, fileName)
	return url, nil
}
