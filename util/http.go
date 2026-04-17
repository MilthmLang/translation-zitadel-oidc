package util

import (
	"bytes"
	"net/http"
)

type ShadowResponseWriter struct {
	header     http.Header
	body       bytes.Buffer
	statusCode int
	maxBodyLen int
}

func NewShadowResponseWriter() *ShadowResponseWriter {
	return &ShadowResponseWriter{
		header:     make(http.Header),
		statusCode: http.StatusOK,
		maxBodyLen: 4 * 1024 * 1024, // 4MB
	}
}

func (s *ShadowResponseWriter) Header() http.Header {
	return s.header
}

func (s *ShadowResponseWriter) Write(bytes []byte) (int, error) {
	remain := s.maxBodyLen - s.body.Len()
	if remain <= 0 {
		return 0, http.ErrContentLength
	}
	if len(bytes) > remain {
		n, err := s.body.Write(bytes[:remain])
		if err != nil {
			return n, err
		}
		return n, http.ErrContentLength
	}
	return s.body.Write(bytes)
}

func (s *ShadowResponseWriter) WriteHeader(statusCode int) {
	s.statusCode = statusCode
}

func (s *ShadowResponseWriter) WriteTo(w http.ResponseWriter) error {
	for k, v := range s.header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.WriteHeader(s.statusCode)

	_, err := w.Write(s.body.Bytes())
	return err
}

func (s *ShadowResponseWriter) StatusCode() int {
	return s.statusCode
}

func (s *ShadowResponseWriter) IsError() bool {
	return s.statusCode >= 400
}

func (s *ShadowResponseWriter) Body() bytes.Buffer {
	return s.body
}
