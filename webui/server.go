package webui

import (
	"encoding/json"
	"github.com/dbratus/loghub/auth"
	"github.com/dbratus/loghub/lhproto"
	"github.com/dbratus/loghub/trace"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const readTimeout = time.Second * 30
const writeTimeout = time.Second * 30
const maxConnections = 30
const dateTimeFormat = "2006-01-02 15:04:05"

var webuiTrace = trace.New("WebUI")

func Start(listenAddr string, certFile string, keyFile string, addr string, useTLS bool, skipCertValidation bool) {
	srv := http.Server{
		Addr:         listenAddr,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/log", logHandler(lhproto.NewClient(addr, maxConnections, useTLS, skipCertValidation)))
	http.HandleFunc("/css/", staticHandler)
	http.HandleFunc("/js/", staticHandler)

	if certFile != "" && keyFile != "" {
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil {
			webuiTrace.Errorf("Failed to start web UI: %s.", err.Error())
		}
	} else {
		if err := srv.ListenAndServe(); err != nil {
			webuiTrace.Errorf("Failed to start web UI: %s.", err.Error())
		}
	}
}

func indexHandler(resp http.ResponseWriter, req *http.Request) {
	if data, err := Asset("content/index.html"); err != nil {
		resp.WriteHeader(404)
	} else {
		resp.Header().Add("Content-Type", "text/html; charset=utf-8")
		resp.Write(data)
	}
}

func staticHandler(resp http.ResponseWriter, req *http.Request) {
	asset := "content" + req.RequestURI

	if data, err := Asset(asset); err != nil {
		resp.WriteHeader(404)
		webuiTrace.Errorf("Failed to get asset: %s.", asset)
	} else {
		if strings.HasSuffix(req.RequestURI, ".js") {
			resp.Header().Add("Content-Type", "text/javascript; charset=utf-8")
		} else if strings.HasSuffix(req.RequestURI, ".css") {
			resp.Header().Add("Content-Type", "text/css; charset=utf-8")
		} else {
			resp.Header().Add("Content-Type", "text/plain; charset=utf-8")
		}

		resp.Write(data)
	}
}

func logHandler(client lhproto.ProtocolHandler) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		fromParam := req.FormValue("from")
		rangeParam := req.FormValue("range")
		minSevParam := req.FormValue("minSev")
		maxSevParam := req.FormValue("maxSev")
		sourcesParam := req.FormValue("sources")
		keywordsParam := req.FormValue("keywords")
		userParam := req.FormValue("user")
		passwordParam := req.FormValue("password")

		webuiTrace.Debugf("Getting log by %s.", req.RequestURI)

		var from, to time.Time

		if t, err := time.ParseInLocation(dateTimeFormat, fromParam, time.Local); err != nil {
			webuiTrace.Errorf("Failed to parse date: %s.", err.Error())
			resp.WriteHeader(500)
			return
		} else {
			from = t
		}

		if r, err := time.ParseDuration(rangeParam); err != nil {
			webuiTrace.Errorf("Failed to parse duration: %s.", err.Error())
			resp.WriteHeader(500)
			return
		} else {
			if r < 0 {
				to = from
				from = to.Add(r)
			} else {
				to = from.Add(r)
			}
		}

		var minSev, maxSev int

		if sev, err := strconv.ParseUint(minSevParam, 10, 8); err != nil {
			webuiTrace.Errorf("Failed to parse severity: %s.", err.Error())
			resp.WriteHeader(500)
			return
		} else {
			minSev = int(sev)
		}

		if sev, err := strconv.ParseUint(maxSevParam, 10, 8); err != nil {
			webuiTrace.Errorf("Failed to parse severity: %s.", err.Error())
			resp.WriteHeader(500)
			return
		} else {
			maxSev = int(sev)
		}

		var sources []string = nil

		if sourcesParam != "" {
			sources = strings.Split(sourcesParam, "\n")
		}

		if userParam == "" {
			userParam = auth.Anonymous
		}

		cred := lhproto.Credentials{userParam, passwordParam}
		queries := make(chan *lhproto.LogQueryJSON)
		results := make(chan *lhproto.OutgoingLogEntryJSON)

		client.Read(&cred, queries, results)

		if sources == nil {
			queries <- &lhproto.LogQueryJSON{from.UnixNano(), to.UnixNano(), minSev, maxSev, ""}
		} else {
			for _, s := range sources {
				queries <- &lhproto.LogQueryJSON{from.UnixNano(), to.UnixNano(), minSev, maxSev, strings.Trim(s, " ")}
			}
		}

		close(queries)

		var keywordsRegEx []*regexp.Regexp = nil

		if keywordsParam != "" {
			keywords := strings.Split(keywordsParam, "\n")
			keywordsRegEx = make([]*regexp.Regexp, len(keywords))

			for i, kw := range keywords {
				if re, err := regexp.Compile(strings.Trim(kw, " ")); err == nil {
					keywordsRegEx[i] = re
				}
			}
		}

		resp.Header().Add("Content-Type", "application/json; charset=utf-8")
		resp.Write([]byte("["))

		webuiTrace.Debugf("Reading results for:\nfrom=%s\nto=%s\nminSev=%d\nmaxSev=%d", from.String(), to.String(), minSev, maxSev)

		isFirst := true
		cnt := 0
		cntReturned := 0

		for ent := range results {
			msgMatch := true

			if keywordsRegEx != nil {
				for _, re := range keywordsRegEx {
					if re != nil && !re.MatchString(ent.Msg) {
						msgMatch = false
					}
				}
			}

			if msgMatch {
				if !isFirst {
					resp.Write([]byte(","))
				}

				if bts, err := json.Marshal(ent); err == nil {
					resp.Write(bts)
					cntReturned++
				}
			}

			isFirst = false
			cnt++
		}

		webuiTrace.Debugf("%d entries read.", cnt)
		webuiTrace.Debugf("%d entries returned.", cntReturned)

		resp.Write([]byte("]"))
	}
}
