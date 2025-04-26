// Copyright (c) 2025 Kevin Berger <huhouhuam@outlook.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package nuget

import (
	"net/http"
)

type ServiceIndex struct {
	Version   string          `json:"version"`
	Resources []*Resource     `json:"resources"`
	Context   *ServiceContext `json:"@context"`
}

type Resource struct {
	Id            string `json:"@id"`
	Type          string `json:"@type"`
	Comment       string `json:"comment"`
	ClientVersion string `json:"clientVersion"`
}

type ServiceContext struct {
	Vocab   string `json:"@vocab"`
	Comment string `json:"comment"`
}

type ServiceResource struct {
	client *Client
}

// GetIndex retrieves the service resources from the NuGet server.
func (s *ServiceResource) GetIndex(options ...RequestOptionFunc) (*ServiceIndex, *http.Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "/v3/index.json", nil, options)
	if err != nil {
		return nil, nil, err
	}
	var svc ServiceIndex
	resp, err := s.client.Do(req, &svc, DecoderTypeJSON)
	if err != nil {
		return nil, resp, err
	}
	return &svc, resp, nil
}
