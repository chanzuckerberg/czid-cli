package czid

import (
	"fmt"
)

type getUserInfoRequest struct {
	SelectedUser struct {
		ProfileFormVersion int `json:"profile_form_version"`
	} `json:"selected_user"`
}

func (c *Client) GetUserInfo(userID int) ([]string, error) {
	var res []string

	err := c.request(
		"GET",
		fmt.Sprintf("/users/%d/edit", userID),
		"",
		getUserInfoRequest{},
		&res,
	)

	return res, err
}
