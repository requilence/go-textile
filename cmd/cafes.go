package cmd

import (
	"net/http"
)

func CafeAdd(url string, token string) error {
	res, err := executeJsonCmd(http.MethodPost, "cafes", params{
		args: []string{url},
		opts: map[string]string{"token": token},
	}, nil)
	if err != nil {
		return err
	}
	output(res)
	return nil
}

func CafeList() error {
	res, err := executeJsonCmd(http.MethodGet, "cafes", params{}, nil)
	if err != nil {
		return err
	}
	output(res)
	return nil
}

func CafeGet(cafeID string) error {
	res, err := executeJsonCmd(http.MethodGet, "cafes/"+cafeID, params{}, nil)
	if err != nil {
		return err
	}
	output(res)
	return nil
}

func CafeDelete(cafeID string) error {
	res, err := executeStringCmd(http.MethodDelete, "cafes/"+cafeID, params{})
	if err != nil {
		return err
	}
	output(res)
	return nil
}

func CafeMessages() error {
	res, err := executeStringCmd(http.MethodGet, "cafes/messages", params{})
	if err != nil {
		return err
	}
	output(res)
	return nil
}
