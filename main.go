package main

import (
	"flag"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/Sirupsen/logrus"
)

var (
	targetAddr        = flag.String("target", "http://localhost:8080", "target destination of proxy server")
	upgradeAddr       = flag.String("u", ":80", "upgrade server address")
	serverAddr        = flag.String("a", ":443", "server address")
	cert              = flag.String("cert", "", "path of certificate file")
	key               = flag.String("key", "", "path of key file")
	stripForwardedFor = flag.Bool("strip-forwarded-for", false, "strip any incoming forwarded for headers before your own")
	formatter         = flag.String("formatter", "text", "logrus formatter, either 'json' or 'text'")
	color             = flag.Bool("color", false, "enable text color for text logs")
)

func init() {
	flag.Parse()
	logrus.SetOutput(os.Stdout)
	switch strings.ToLower(*formatter) {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{
			DisableColors: !*color,
		})
	}
}

func handleUpgrade(w http.ResponseWriter, r *http.Request) {
	newurl := "https://" + r.Host + ":" + r.URL.String()
	http.Redirect(w, r, newurl, 302)
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// stripPort strips the port from an IP
func stripPort(ip string) string {
	return strings.TrimSpace(strings.Split(ip, ":")[0])
}

// printReq prints a request to stderr
func printReq(r *http.Request) {
	logrus.WithFields(logrus.Fields{
		"referer":    r.Referer(),
		"address":    stripPort(r.RemoteAddr),
		"host":       r.Host,
		"url":        r.URL.String(),
		"user-agent": r.UserAgent(),
		"method":     r.Method,
	}).Info(stripPort(r.RemoteAddr), " ", r.URL.String(), " ", r.Method, " ", r.UserAgent())
}

// NewReverseProxy returns a new reverse proxy
func NewReverseProxy(target *url.URL) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {

		// Set the X-Forwarded-For header so the receiver knows where the request came from
		// If stripForwardedFor is true and a previous header is found, remove it and replace
		// It with a single address.
		h, ok := req.Header["X-Forwarded-For"]
		ip := stripPort(req.RemoteAddr)
		if ok && !*stripForwardedFor {
			req.Header.Set("X-Forwarded-For", h[0]+", "+ip)
		} else {
			req.Header.Set("X-Forwarded-For", ip)
		}

		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		printReq(req)
	}
	return &httputil.ReverseProxy{Director: director}
}

func main() {
	targetURL, err := url.Parse(*targetAddr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Failed to parse targer URL", err)
	}

	m := http.NewServeMux()
	u := http.NewServeMux()

	m.Handle("/", NewReverseProxy(targetURL))
	u.HandleFunc("/", handleUpgrade)

	srvUpgrade := http.Server{
		Addr:    *upgradeAddr,
		Handler: u,
	}

	srvTLS := http.Server{
		Addr:    *serverAddr,
		Handler: m,
	}

	if *upgradeAddr == "" && *serverAddr == "" {
		logrus.Fatal("Neither an upgrade server address, or server address exists, exiting...")
	}

	if *upgradeAddr != "" {
		go func() {
			err := srvUpgrade.ListenAndServe()
			if err != nil {
				logrus.WithField("err", err).Fatal("error hosting upgrade server: ", err)
			}
		}()
	}
	if *serverAddr != "" {
		go func() {
			var err error
			if *key != "" && *cert != "" {
				err = srvTLS.ListenAndServeTLS(*cert, *key)
			} else {
				logrus.WithFields(logrus.Fields{
					"cert": *cert,
					"key":  *key,
				}).Warn("Either a key file or certificate file was missing from the flags, serving HTTP")
				err = srvTLS.ListenAndServe()
			}
			if err != nil {
				logrus.WithField("err", err).Fatal("error hosting TLS server: ", err)
			}
		}()
	}

	s := make(chan os.Signal)
	signal.Notify(s, os.Kill, os.Interrupt)
	<-s
	logrus.Info("Received sigterm or sigkill, exiting...")
}
