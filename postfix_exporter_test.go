package main

import (
	"github.com/hpcloud/tail"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPostfixExporter_CollectFromLogline(t *testing.T) {
	type fields struct {
		showqPath                       string
		journal                         *Journal
		tailer                          *tail.Tail
		cleanupProcesses                prometheus.Counter
		cleanupRejects                  prometheus.Counter
		cleanupNotAccepted              prometheus.Counter
		lmtpDelays                      *prometheus.HistogramVec
		pipeDelays                      *prometheus.HistogramVec
		qmgrInsertsNrcpt                prometheus.Histogram
		qmgrInsertsSize                 prometheus.Histogram
		qmgrRemoves                     prometheus.Counter
		smtpDelays                      *prometheus.HistogramVec
		smtpTLSConnects                 *prometheus.CounterVec
		smtpDeferreds                   prometheus.Counter
		smtpdConnects                   prometheus.Counter
		smtpdDisconnects                prometheus.Counter
		smtpdFCrDNSErrors               prometheus.Counter
		smtpdLostConnections            *prometheus.CounterVec
		smtpdProcesses                  *prometheus.CounterVec
		smtpdRejects                    *prometheus.CounterVec
		smtpdSASLAuthenticationFailures prometheus.Counter
		smtpdTLSConnects                *prometheus.CounterVec
		unsupportedLogEntries           *prometheus.CounterVec
		postscreenRejects               *prometheus.CounterVec
	}
	type args struct {
		line              []string
		removedCount      int
		saslFailedCount   int
		outgoingTLS       int
		postscreenRejects int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Single line",
			args: args{
				line: []string{
					"Feb 11 16:49:24 letterman postfix/qmgr[8204]: AAB4D259B1: removed",
				},
				removedCount:    1,
				saslFailedCount: 0,
			},
			fields: fields{
				qmgrRemoves:           &testCounter{count: 0},
				unsupportedLogEntries: prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"process"}),
			},
		},
		{
			name: "Multiple lines",
			args: args{
				line: []string{
					"Feb 11 16:49:24 letterman postfix/qmgr[8204]: AAB4D259B1: removed",
					"Feb 11 16:49:24 letterman postfix/qmgr[8204]: C2032259E6: removed",
					"Feb 11 16:49:24 letterman postfix/qmgr[8204]: B83C4257DC: removed",
					"Feb 11 16:49:24 letterman postfix/qmgr[8204]: 721BE256EA: removed",
					"Feb 11 16:49:25 letterman postfix/qmgr[8204]: CA94A259EB: removed",
					"Feb 11 16:49:25 letterman postfix/qmgr[8204]: AC1E3259E1: removed",
					"Feb 11 16:49:25 letterman postfix/qmgr[8204]: D114D221E3: removed",
					"Feb 11 16:49:25 letterman postfix/qmgr[8204]: A55F82104D: removed",
					"Feb 11 16:49:25 letterman postfix/qmgr[8204]: D6DAA259BC: removed",
					"Feb 11 16:49:25 letterman postfix/qmgr[8204]: E3908259F0: removed",
					"Feb 11 16:49:25 letterman postfix/qmgr[8204]: 0CBB8259BF: removed",
					"Feb 11 16:49:25 letterman postfix/qmgr[8204]: EA3AD259F2: removed",
					"Feb 11 16:49:25 letterman postfix/qmgr[8204]: DDEF824B48: removed",
					"Feb 11 16:49:26 letterman postfix/qmgr[8204]: 289AF21DB9: removed",
					"Feb 11 16:49:26 letterman postfix/qmgr[8204]: 6192B260E8: removed",
					"Feb 11 16:49:26 letterman postfix/qmgr[8204]: F2831259F4: removed",
					"Feb 11 16:49:26 letterman postfix/qmgr[8204]: 09D60259F8: removed",
					"Feb 11 16:49:26 letterman postfix/qmgr[8204]: 13A19259FA: removed",
					"Feb 11 16:49:26 letterman postfix/qmgr[8204]: 2D42722065: removed",
					"Feb 11 16:49:26 letterman postfix/qmgr[8204]: 746E325A0E: removed",
					"Feb 11 16:49:26 letterman postfix/qmgr[8204]: 4D2F125A02: removed",
					"Feb 11 16:49:26 letterman postfix/qmgr[8204]: E30BC259EF: removed",
					"Feb 11 16:49:26 letterman postfix/qmgr[8204]: DC88924DA1: removed",
					"Feb 11 16:49:26 letterman postfix/qmgr[8204]: 2164B259FD: removed",
					"Feb 11 16:49:26 letterman postfix/qmgr[8204]: 8C30525A14: removed",
					"Feb 11 16:49:26 letterman postfix/qmgr[8204]: 8DCCE25A15: removed",
					"Feb 11 16:49:26 letterman postfix/qmgr[8204]: C5217255D5: removed",
					"Feb 11 16:49:27 letterman postfix/qmgr[8204]: D8EE625A28: removed",
					"Feb 11 16:49:27 letterman postfix/qmgr[8204]: 9AD7C25A19: removed",
					"Feb 11 16:49:27 letterman postfix/qmgr[8204]: D0EEE2596C: removed",
					"Feb 11 16:49:27 letterman postfix/qmgr[8204]: DFE732172E: removed",
				},
				removedCount:    31,
				saslFailedCount: 0,
			},
			fields: fields{
				qmgrRemoves:           &testCounter{count: 0},
				unsupportedLogEntries: prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"process"}),
			},
		},
		{
			name: "SASL Failed",
			args: args{
				line: []string{
					"Apr 26 10:55:19 tcc1 postfix/smtpd[21126]: warning: SASL authentication failure: cannot connect to saslauthd server: Permission denied",
					"Apr 26 10:55:19 tcc1 postfix/smtpd[21126]: warning: SASL authentication failure: Password verification failed",
					"Apr 26 10:55:19 tcc1 postfix/smtpd[21126]: warning: laptop.local[192.168.1.2]: SASL PLAIN authentication failed: generic failure",
				},
				saslFailedCount: 1,
				removedCount:    0,
			},
			fields: fields{
				smtpdSASLAuthenticationFailures: &testCounter{count: 0},
				unsupportedLogEntries:           prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"process"}),
			},
		},
		{
			name: "Issue #35",
			args: args{
				line: []string{
					"Jul 24 04:38:17 mail postfix/smtp[30582]: Verified TLS connection established to gmail-smtp-in.l.google.com[108.177.14.26]:25: TLSv1.3 with cipher TLS_AES_256_GCM_SHA384 (256/256 bits) key-exchange X25519 server-signature RSA-PSS (2048 bits) server-digest SHA256",
					"Jul 24 03:28:15 mail postfix/smtp[24052]: Verified TLS connection established to mx2.comcast.net[2001:558:fe21:2a::6]:25: TLSv1.2 with cipher ECDHE-RSA-AES256-GCM-SHA384 (256/256 bits)",
				},
				removedCount:    0,
				saslFailedCount: 0,
				outgoingTLS:     2,
			},
			fields: fields{
				unsupportedLogEntries: prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"process"}),
				smtpTLSConnects:       prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"Verified", "TLSv1.2", "ECDHE-RSA-AES256-GCM-SHA384", "256", "256"}),
			},
		},
		{
			name: "Issue #36",
			args: args{
				line: []string{
					"Feb 22 03:18:19 <hostname> postfix/postscreen[1234]: WHITELISTED [1.2.3.4]:12345",
					"Feb 22 03:20:57 <hostname> postfix/postscreen[1234]: NOQUEUE: reject: RCPT from [1.2.3.4]:12345: 550 5.7.1 Service unavailable; client [<spammers ip>] blocked using DNSBL Filters; from=<fromaddr>, to=<toaddr>, proto=ESMTP, helo=<smtp.aweia.cn>",
					"Nov 22 16:03:56 siren postfix/postscreen[8266]: NOQUEUE: reject: RCPT from [209.85.160.43]:45612: 450 4.3.2 Service currently unavailable; from=account@gmail.com, to=user@abc.com, proto=ESMTP, helo=<mail-pl0-f43.google.com>",
				},
				removedCount:      0,
				saslFailedCount:   0,
				outgoingTLS:       0,
				postscreenRejects: 2,
			},
			fields: fields{
				unsupportedLogEntries: prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"process"}),
				smtpTLSConnects:       prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"Verified", "TLSv1.2", "ECDHE-RSA-AES256-GCM-SHA384", "256", "256"}),
				postscreenRejects:     prometheus.NewCounterVec(prometheus.CounterOpts{}, []string{"code"}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &PostfixExporter{
				showqPath:                       tt.fields.showqPath,
				journal:                         tt.fields.journal,
				tailer:                          tt.fields.tailer,
				cleanupProcesses:                tt.fields.cleanupProcesses,
				cleanupRejects:                  tt.fields.cleanupRejects,
				cleanupNotAccepted:              tt.fields.cleanupNotAccepted,
				lmtpDelays:                      tt.fields.lmtpDelays,
				pipeDelays:                      tt.fields.pipeDelays,
				qmgrInsertsNrcpt:                tt.fields.qmgrInsertsNrcpt,
				qmgrInsertsSize:                 tt.fields.qmgrInsertsSize,
				qmgrRemoves:                     tt.fields.qmgrRemoves,
				smtpDelays:                      tt.fields.smtpDelays,
				smtpTLSConnects:                 tt.fields.smtpTLSConnects,
				smtpDeferreds:                   tt.fields.smtpDeferreds,
				smtpdConnects:                   tt.fields.smtpdConnects,
				smtpdDisconnects:                tt.fields.smtpdDisconnects,
				smtpdFCrDNSErrors:               tt.fields.smtpdFCrDNSErrors,
				smtpdLostConnections:            tt.fields.smtpdLostConnections,
				smtpdProcesses:                  tt.fields.smtpdProcesses,
				smtpdRejects:                    tt.fields.smtpdRejects,
				smtpdSASLAuthenticationFailures: tt.fields.smtpdSASLAuthenticationFailures,
				smtpdTLSConnects:                tt.fields.smtpdTLSConnects,
				unsupportedLogEntries:           tt.fields.unsupportedLogEntries,
				postscreenRejects:               tt.fields.postscreenRejects,
				logUnsupportedLines:             true,
			}
			for _, line := range tt.args.line {
				e.CollectFromLogLine(line)
			}
			assertCounterEquals(t, e.qmgrRemoves, tt.args.removedCount, "Wrong number of lines counted")
			assertCounterEquals(t, e.smtpdSASLAuthenticationFailures, tt.args.saslFailedCount, "Wrong number of Sasl counter counted")
			assertCounterVecEquals(t, e.smtpTLSConnects, tt.args.outgoingTLS, "Wrong number of TLS connections counted")
			assertCounterVecEquals(t, e.postscreenRejects, tt.args.postscreenRejects, "Wrong number of messages rejected")
		})
	}
}
func assertCounterVecEquals(t *testing.T, counter prometheus.Collector, expected int, message string) {

	if counter != nil && expected > 0 {
		switch counter.(type) {
		case *prometheus.CounterVec:
			counter := counter.(*prometheus.CounterVec)
			metricsChan := make(chan prometheus.Metric)
			go func() {
				counter.Collect(metricsChan)
				close(metricsChan)
			}()
			var count int = 0
			for metric := range metricsChan {
				metricDto := io_prometheus_client.Metric{}
				metric.Write(&metricDto)
				count += int(*metricDto.Counter.Value)
			}
			assert.Equal(t, expected, count, message)
		default:
			t.Fatal("Type not implemented")
		}
	}
}
func assertCounterEquals(t *testing.T, counter prometheus.Counter, expected int, message string) {

	if counter != nil && expected > 0 {
		switch counter.(type) {
		case *testCounter:
			counter := counter.(*testCounter)
			assert.Equal(t, expected, counter.Count(), message)
		default:
			t.Fatal("Type not implemented")
		}
	}
}

type testCounter struct {
	count int
}

func (t *testCounter) setCount(count int) {
	t.count = count
}

func (t *testCounter) Count() int {
	return t.count
}

func (t *testCounter) Add(_ float64) {
}
func (t *testCounter) Collect(_ chan<- prometheus.Metric) {
}
func (t *testCounter) Describe(_ chan<- *prometheus.Desc) {
}
func (t *testCounter) Desc() *prometheus.Desc {
	return nil
}
func (t *testCounter) Inc() {
	t.count++
}
func (t *testCounter) Write(_ *io_prometheus_client.Metric) error {
	return nil
}
