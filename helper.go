package main

import (
	"errors"
	"net/mail"
	"strings"
)

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
