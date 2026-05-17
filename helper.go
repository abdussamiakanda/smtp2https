package main

import (
	"encoding/json"
	"errors"
	"net/mail"
	"strings"
)

func mergeWebhookPayload(email EmailMessage, extra map[string]interface{}) (map[string]interface{}, error) {
	raw, err := json.Marshal(email)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	for key, value := range extra {
		payload[key] = value
	}

	return payload, nil
}

func recipientDomain(address string) (string, error) {
	addr, err := mail.ParseAddress(strings.TrimSpace(address))
	if err != nil {
		at := strings.LastIndex(address, "@")
		if at < 0 || at == len(address)-1 {
			return "", errors.New("invalid recipient address")
		}
		return normalizeDomain(address[at+1:]), nil
	}

	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 || parts[1] == "" {
		return "", errors.New("invalid recipient address")
	}

	return normalizeDomain(parts[1]), nil
}

func transformStdAddressToEmailAddress(addr []*mail.Address) []*EmailAddress {
	ret := []*EmailAddress{}

	for _, e := range addr {
		ret = append(ret, &EmailAddress{
			Address: e.Address,
			Name:    e.Name,
		})
	}

	return ret
}

// func smtpsrvMesssage2EmailMessage(msg *smtpsrv.Context)
