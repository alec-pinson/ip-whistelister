package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

/*
- name
- employee_id
- ip
- func new
- func whitelist
- func unwhitelist
*/

type User struct {
	key        string
	name       string
	employeeId string
	ip         string
	cidr       string // microsoft saying without /32 can cause issues... dont believe them but w/e ticket id - 2106010050001687
}

func (u *User) new(client *http.Client, req *http.Request) *User {
	// get display name + employee id
	resp, err := client.Get("https://graph.windows.net/me?api-version=1.6")
	if err != nil {
		log.Printf("user.new(): error creating token  %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("user.new(): token response was %s", resp.Status)
		return nil
	}

	var ud map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&ud); err != nil {
		log.Printf("user.new(): error decoding JSON response: %v", err)
		return nil
	}

	if c.Debug {
		log.Printf("user.new(): %v", ud)
	}

	u.employeeId = fmt.Sprintf("%v", ud["employeeId"])
	u.name = fmt.Sprintf("%v", ud["displayName"])

	// Create our 'key' by removing spaces, converting to lower and removing all special characters
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	u.key = strings.ToLower(reg.ReplaceAllString(u.name+u.employeeId, ""))

	// get ip
	u.ip = req.Header.Get("X-Azure-Clientip")
	if u.ip == "" {
		u.ip, _, err = net.SplitHostPort(req.RemoteAddr)
		if err != nil {
			log.Printf("user.new(): %q is not IP:port\n", req.RemoteAddr)
		}
	}

	u.cidr = u.ip + "/32"

	log.Println("user.new(): authentication successful - " + u.name + " (" + u.employeeId + ") - " + u.ip)

	return u
}

func (u *User) whitelist() {
	s := w.add(u)
	if s {
		log.Println("user.whitelist(): Whitelisting for '" + u.ip + "' (" + u.name + ") will expire on " + time.Now().Add(whitelistTTL*time.Hour).Format("02-01-2006 at 15:04"))
	}
}

func (u *User) unwhitelist() {
	s := w.delete(u)
	if s {
		log.Println("user.unwhitelist(): Whitelisting for '" + u.ip + "' (" + u.name + ") has been removed")
	}
}