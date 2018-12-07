package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

func badRequest(resp http.ResponseWriter, message string) {
	http.Error(resp, message, http.StatusBadRequest)
}

func internalError(resp http.ResponseWriter, message string) {
	http.Error(resp, message, http.StatusInternalServerError)
}

func readBody(reader io.Reader) string {
	bytes, err := ioutil.ReadAll(reader)
	if err != nil || len(bytes) == 0 {
		return ""
	}
	return string(bytes)
}

func readBodyJson(reader io.Reader, obj interface{}) error {
	text := readBody(reader)
	return json.Unmarshal([]byte(text), obj)
}

func writeBody(writer io.Writer, data string) {
	fmt.Fprint(writer, data)
}

func writeBodyJson(resp http.ResponseWriter, obj interface{}) error {
	buff, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	resp.Header().Set("Content-Type", "application/json; charset=utf-8")
	writedLen, err := resp.Write(buff)
	if err != nil {
		return err
	}
	if writedLen != len(buff) {
		return errors.New("Not all data writed")
	}
	return nil
}
