package ldaphandler

import (
	"crypto/tls"
	"errors"
	"fmt"
	"gopkg.in/ldap.v2"
)

type LdapClient struct {
	Address            string
	Base               string
	BindDN             string
	BindPassword       string
	GroupFilter        string // e.g. "(memberUid=%s)"
	ServerName         string
	UserFilter         string // e.g. "(uid=%s)"
	Conn               *ldap.Conn
	InsecureSkipVerify bool
	UseSSL             bool
	SkipTLS            bool
	ClientCertificates []tls.Certificate // Adding client certificates
}

func NewClient(config LdapConfig) (*LdapClient, error) {
	client := LdapClient{
		Address:            config.Address,
		Base:               config.Base,
		BindDN:             config.Bind.User,
		BindPassword:       config.Bind.Password,
		GroupFilter:        config.Filter.Group,
		ServerName:         config.Ssl.ServerName,
		UserFilter:         config.Filter.User,
		Conn:               nil,
		InsecureSkipVerify: config.Ssl.SkipVerify,
		UseSSL:             config.Ssl.Use,
		SkipTLS:            config.Ssl.SkipTls,
		ClientCertificates: nil,
	}
	return &client, client.Connect()
}

func (lc *LdapClient) IsAlive() bool {
	return lc.Conn != nil
}

// Connect connects to the ldap backend.
func (lc *LdapClient) Connect() error {
	if lc.Conn == nil {
		var l *ldap.Conn
		var err error
		if !lc.UseSSL {
			l, err = ldap.Dial("tcp", lc.Address)
			if err != nil {
				return err
			}
			// Reconnect with TLS
			if !lc.SkipTLS {
				err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
				if err != nil {
					return err
				}
			}
		} else {
			config := &tls.Config{
				InsecureSkipVerify: lc.InsecureSkipVerify,
				ServerName:         lc.ServerName,
			}
			if lc.ClientCertificates != nil && len(lc.ClientCertificates) > 0 {
				config.Certificates = lc.ClientCertificates
			}
			l, err = ldap.DialTLS("tcp", lc.Address, config)
			if err != nil {
				return err
			}
		}

		lc.Conn = l
	}
	return nil
}

// Close closes the ldap backend connection.
func (lc *LdapClient) Close() {
	if lc.Conn != nil {
		lc.Conn.Close()
	}
	lc.Conn = nil
}

// Authenticate authenticates the user against the ldap backend.
func (lc *LdapClient) Authenticate(username, password string) (bool, map[string][]string, error) {
	err := lc.Connect()
	if err != nil {
		return false, nil, err
	}

	// First bind with a read only user
	if lc.BindDN != "" && lc.BindPassword != "" {
		err := lc.Conn.Bind(lc.BindDN, lc.BindPassword)
		if err != nil {
			if err.Error() == `LDAP Result Code 200 "Network Error": ldap: connection closed` {
				lc.Close()
				return lc.Authenticate(username, password)
			}
			return false, nil, err
		}
	}

	//attributes := append(lc.Attributes, "mail")
	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		lc.Base,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf(lc.UserFilter, username),
		nil,
		nil,
	)

	searchResult, err := lc.Conn.Search(searchRequest)
	if err != nil {
		return false, nil, err
	}

	if len(searchResult.Entries) < 1 {
		return false, nil, errors.New("user does not exist")
	}

	if len(searchResult.Entries) > 1 {
		return false, nil, errors.New("too many entries returned")
	}

	userData := map[string][]string{}
	for _, attr := range searchResult.Entries[0].Attributes {
		userData[attr.Name] = attr.Values
	}

	// Bind as the user to verify their password
	err = lc.Conn.Bind(searchResult.Entries[0].DN, password)
	if err != nil {
		return false, userData, err
	}

	// Rebind as the read only user for any further queries
	if lc.BindDN != "" && lc.BindPassword != "" {
		err = lc.Conn.Bind(lc.BindDN, lc.BindPassword)
		if err != nil {
			return true, userData, err
		}
	}

	return true, userData, nil
}

// GetGroupsOfUser returns the group for a user.
func (lc *LdapClient) GetGroupsOfUser(username string) ([]string, error) {
	err := lc.Connect()
	if err != nil {
		return nil, err
	}

	searchRequest := ldap.NewSearchRequest(
		lc.Base,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf(lc.GroupFilter, username),
		[]string{"cn"}, // can it be something else than "cn"?
		nil,
	)
	sr, err := lc.Conn.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	var groups []string
	for _, entry := range sr.Entries {
		groups = append(groups, entry.GetAttributeValue("cn"))
	}
	return groups, nil
}
