package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
)

type Rule struct {
	Endpoint               string   `yaml:"endpoint"`
	ForbiddenUserAgents    []string `yaml:"forbidden_user_agents"`
	ForbiddenHeaders       []string `yaml:"forbidden_headers"`
	RequiredHeaders        []string `yaml:"required_headers"`
	MaxRequestLengthBytes  int      `yaml:"max_request_length_bytes"`
	MaxResponseLengthBytes int      `yaml:"max_response_length_bytes"`
	ForbiddenResponseCodes []int    `yaml:"forbidden_response_codes"`
	ForbiddenRequestRe     []string `yaml:"forbidden_request_re"`
	ForbiddenResponseRe    []string `yaml:"forbidden_response_re"`
}

type Config struct {
	Rules []Rule `yaml:"rules"`
}

var (
	serviceAddr = flag.String("service-addr", "", "address of protected service")
	confPath    = flag.String("conf", "", "path to config file")
	addr        = flag.String("addr", "", "address to listen on")
	config      Config
)

func main() {
	flag.Parse()

	if err := loadConfig(*confPath); err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	targetURL, err := url.Parse(*serviceAddr)
	if err != nil {
		fmt.Printf("Error parsing service address: %v\n", err)
		os.Exit(1)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = &customTransport{}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		for _, rule := range config.Rules {
			if rule.Endpoint == "" || rule.Endpoint == r.URL.Path {
				if checkForbiddenRequest(w, r, rule) {
					return
				}
			}
		}

		proxy.ServeHTTP(w, r)
	})

	fmt.Printf("Firewall listening on %s, protecting %s\n", *addr, *serviceAddr)
	http.ListenAndServe(*addr, nil)
}

func checkForbiddenRequest(w http.ResponseWriter, r *http.Request, rule Rule) bool {
	if checkUserAgent(w, r, rule) || checkForbiddenHeaders(w, r, rule) ||
		checkRequiredHeaders(w, r, rule) || checkRequestLength(w, r, rule) ||
		checkForbiddenRequestRe(w, r, rule) {
		return true
	}
	return false
}

func checkUserAgent(w http.ResponseWriter, r *http.Request, rule Rule) bool {
	for _, ua := range rule.ForbiddenUserAgents {
		isUAForbidden, _ := regexp.MatchString(ua, r.UserAgent())
		if isUAForbidden {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return true
		}
	}
	return false
}

func checkForbiddenHeaders(w http.ResponseWriter, r *http.Request, rule Rule) bool {
	for _, forbiddenHeader := range rule.ForbiddenHeaders {
		headerParts := strings.SplitN(forbiddenHeader, ":", 2)
		if len(headerParts) == 2 {
			key := strings.TrimSpace(headerParts[0])
			value := strings.TrimSpace(headerParts[1])
			if r.Header.Get(key) == value {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return true
			}
		}
	}
	return false
}

func checkRequiredHeaders(w http.ResponseWriter, r *http.Request, rule Rule) bool {
	for _, requiredHeader := range rule.RequiredHeaders {
		if r.Header.Get(requiredHeader) == "" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return true
		}
	}
	return false
}

func checkRequestLength(w http.ResponseWriter, r *http.Request, rule Rule) bool {
	if rule.MaxRequestLengthBytes > 0 && r.ContentLength > int64(rule.MaxRequestLengthBytes) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return true
	}
	return false
}

func checkForbiddenRequestRe(w http.ResponseWriter, r *http.Request, rule Rule) bool {
	if len(rule.ForbiddenRequestRe) > 0 {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return true
		}
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		for _, pattern := range rule.ForbiddenRequestRe {
			matches, _ := regexp.MatchString(pattern, string(body))
			if matches {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return true
			}
		}
	}
	return false
}

func loadConfig(path string) error {
	configFile, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(configFile, &config)
}

type customTransport struct{}

func (t *customTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	for _, rule := range config.Rules {
		if rule.Endpoint == "" || rule.Endpoint == r.URL.Path {
			if checkForbiddenResponseCodes(response, rule) || checkResponseLength(response, rule) || checkForbiddenResponseRe(response, rule) {
				return nil, fmt.Errorf("response blocked")
			}
		}
	}

	return response, nil
}

func checkForbiddenResponseCodes(response *http.Response, rule Rule) bool {
	for _, forbiddenCode := range rule.ForbiddenResponseCodes {
		if response.StatusCode == forbiddenCode {
			response.Body.Close()
			return true
		}
	}
	return false
}

func checkResponseLength(response *http.Response, rule Rule) bool {
	if rule.MaxResponseLengthBytes > 0 && response.ContentLength > int64(rule.MaxResponseLengthBytes) {
		response.Body.Close()
		return true
	}
	return false
}

func checkForbiddenResponseRe(response *http.Response, rule Rule) bool {
	if len(rule.ForbiddenResponseRe) > 0 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			response.Body.Close()
			return true
		}
		response.Body = io.NopCloser(strings.NewReader(string(body)))

		for _, pattern := range rule.ForbiddenResponseRe {
			matches, _ := regexp.MatchString(pattern, string(body))
			if matches {
				response.Body.Close()
				return true
			}
		}
	}
	return false
}
