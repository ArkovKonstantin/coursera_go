package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type Root struct {
	XMLName xml.Name `xml:"root"`
	Text    string   `xml:",chardata"`
	Row     []struct {
		Text          string `xml:",chardata"`
		ID            int    `xml:"id"`
		Guid          string `xml:"guid"`
		IsActive      string `xml:"isActive"`
		Balance       string `xml:"balance"`
		Picture       string `xml:"picture"`
		Age           int    `xml:"age"`
		EyeColor      string `xml:"eyeColor"`
		FirstName     string `xml:"first_name"`
		LastName      string `xml:"last_name"`
		Gender        string `xml:"gender"`
		Company       string `xml:"company"`
		Email         string `xml:"email"`
		Phone         string `xml:"phone"`
		Address       string `xml:"address"`
		About         string `xml:"about"`
		Registered    string `xml:"registered"`
		FavoriteFruit string `xml:"favoriteFruit"`
	} `xml:"row"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	orderFieldValues := map[string]bool{"Id": true, "Age": true, "Name": true}
	orderFiled := r.URL.Query().Get("order_field")
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	accessToken := r.Header.Get("AccessToken")

	if accessToken != "token" {
		w.WriteHeader(http.StatusUnauthorized)
	} else if _, ok := orderFieldValues[orderFiled]; orderFiled != "" && !ok {
		w.WriteHeader(http.StatusBadRequest)
		b, _ := json.Marshal(SearchErrorResponse{"ErrorBadOrderField"})
		w.Write(b)

	} else {
		f, _ := os.Open("dataset.xml")
		root := Root{}
		users := []User{}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			log.Fatal(err)
		}
		_ = xml.Unmarshal(b, &root)
		for _, r := range root.Row {
			users = append(users, User{
				r.ID,
				r.FirstName + r.LastName,
				r.Age,
				r.About,
				r.Gender,
			})
		}
		//fmt.Println("len", len(users))
		limit = offset + limit
		if limit > len(users) {
			limit = len(users)
		}
		data, err := json.Marshal(users[offset:limit])
		if err != nil {
			log.Fatal(err)
		}
		w.Write(data)
	}
}

func TestNegativeLimit(t *testing.T) {
	req := SearchRequest{Limit: -1}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	client := SearchClient{AccessToken: "token", URL: ts.URL}
	_, err := client.FindUsers(req)
	if err == nil || err.Error() != "limit must be > 0" {
		t.Error("the limit parameter is not checked")
	}
}

func TestNegativeOffset(t *testing.T) {
	req := SearchRequest{Offset: -1}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	client := SearchClient{AccessToken: "token", URL: ts.URL}
	_, err := client.FindUsers(req)
	if err == nil || err.Error() != "offset must be > 0" {
		t.Error("the offset parameter is not checked")
	}
}
func TestInvalidOrderField(t *testing.T) {
	req := SearchRequest{OrderField: "Invalid"}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	client := SearchClient{AccessToken: "token", URL: ts.URL}
	_, err := client.FindUsers(req)
	if err == nil || err.Error() != fmt.Sprintf("OrderFeld %s invalid", req.OrderField) {
		//fmt.Println(err.Error())
		t.Error("the order_field parameter is not checked")
	}
}
func TestGettingFirstPage(t *testing.T) {
	req := SearchRequest{25, 0, "", "", 0}
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	client := SearchClient{AccessToken: "token", URL: ts.URL}

	resp, _ := client.FindUsers(req)
	if len(resp.Users) != 25 {
		t.Error("num of users must be equal limit")
	}
	if !resp.NextPage {
		t.Error("Next page must be true")
	}

}

func TestAccessToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	client := SearchClient{AccessToken: "badToken", URL: ts.URL}
	_, err := client.FindUsers(SearchRequest{})
	if err == nil || err.Error() != "Bad AccessToken" {
		t.Error("AccessToken is not checked")
	}
}

func TestServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	client := SearchClient{URL: ts.URL}
	_, err := client.FindUsers(SearchRequest{})
	if err == nil || err.Error() != "SearchServer fatal error" {
		t.Error("internal server error c")
	}
}

func TestUnknownBadRequestError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		b, _ := json.Marshal(SearchErrorResponse{Error: "WrongErrMsg"})
		w.Write(b)
	}))
	client := SearchClient{URL: ts.URL}
	_, err := client.FindUsers(SearchRequest{})
	if err == nil || err.Error() != "unknown bad request error: WrongErrMsg" {
		t.Error("unknown bad req error is not handled")
	}
}

func TestUnpackBadRequestError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	client := SearchClient{URL: ts.URL}
	_, err := client.FindUsers(SearchRequest{})
	if err == nil || err.Error() != "cant unpack error json: unexpected end of JSON input" {
		t.Error("unpack bad req error is not handled")
	}
}

func TestUnpackResultError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte{})
	}))
	client := SearchClient{URL: ts.URL}
	_, err := client.FindUsers(SearchRequest{})
	//fmt.Println(err)
	if err == nil || err.Error() != "cant unpack result json: unexpected end of JSON input" {
		t.Error("unpack result error is not handled")
	}
}

func TestTimeoutError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
	}))
	client := SearchClient{URL: ts.URL}
	_, err := client.FindUsers(SearchRequest{})
	//fmt.Println(err)
	if err == nil || err.Error() != "timeout for limit=1&offset=0&order_by=0&order_field=&query=" {
		t.Error("timeout error is not handled")
	}
}

func TestUnknownError(t *testing.T) {
	client := SearchClient{URL: "no_valid_url"}
	_, err := client.FindUsers(SearchRequest{})
	if err == nil || !strings.HasPrefix(err.Error(), "unknown error") {
		t.Error("unknown error is not handled")
	}
}

func TestGettingLastPage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	req := SearchRequest{100, 10, "", "", 0}
	client := SearchClient{"token", ts.URL}
	resp, _ := client.FindUsers(req)
	//fmt.Println(len(resp.Users))
	//fmt.Println(resp.NextPage)
	if len(resp.Users) != 25 {
		t.Error("must return last 25 users")
	}
	if resp.NextPage {
		t.Error("Next page must be false")
	}
}
