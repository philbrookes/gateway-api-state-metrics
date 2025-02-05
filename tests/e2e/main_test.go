package metrics

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	ksmFramework "k8s.io/kube-state-metrics/v2/tests/e2e/framework"
)

var framework *ksmFramework.Framework

func TestMain(m *testing.M) {
	ksmHTTPMetricsURL := flag.String(
		"ksm-http-metrics-url",
		"",
		"url to access the kube-state-metrics service",
	)
	ksmTelemetryURL := flag.String(
		"ksm-telemetry-url",
		"",
		"url to access the kube-state-metrics telemetry endpoint",
	)
	flag.Parse()

	var (
		err      error
		exitCode int
	)

	if framework, err = ksmFramework.New(*ksmHTTPMetricsURL, *ksmTelemetryURL); err != nil {
		log.Fatalf("failed to setup framework: %v\n", err)
	}

	exitCode = m.Run()

	os.Exit(exitCode)
}

func TestGatewayMetricsAvailable(t *testing.T) {
	buf := &bytes.Buffer{}

	err := framework.KsmClient.Metrics(buf)
	if err != nil {
		t.Fatalf("failed to get metrics from kube-state-metrics: %v", err)
	}

	// Ideally we could use framework.ParseMetrics here,
	// however it gives an error like this:
	// Failed to get or decode telemetry metrics text format parsing error in line 1231: unknown metric type "info"
	// The underlying parsing library doesn't seem to allow OpenMetrics format
	// and works with just Prometheus format
	// Related issues where I found some info on this:
	// - https://github.com/prometheus/pushgateway/issues/400
	// - https://github.com/prometheus/pushgateway/issues/479
	// - https://discuss.elastic.co/t/using-kube-state-metrics-custom-resource-state-metrics-breaks-metricbeat/341249

	re := regexp.MustCompile(`^(gatewayapi_.*){(.*)}\s+(.*)`)
	scanner := bufio.NewScanner(buf)
	gatewayapiMetrics := map[string][][]string{}
	for scanner.Scan() {
		// fmt.Printf("checking metric text=%s\n", scanner.Text())
		params := re.FindStringSubmatch(scanner.Text())
		// fmt.Printf("params=%v\n", params)
		if len(params) < 4 {
			continue
		}
		if gatewayapiMetrics[params[1]] == nil {
			gatewayapiMetrics[params[1]] = [][]string{}
		}
		fmt.Printf("Adding matched metric params=%v\n", params)
		gatewayapiMetrics[params[1]] = append(gatewayapiMetrics[params[1]], params)
	}
	testGatewayClasses(t, gatewayapiMetrics)
	testGateways(t, gatewayapiMetrics)
	testHTTPRoutes(t, gatewayapiMetrics)
	testGRPCRoutes(t, gatewayapiMetrics)
	testTCPRoute(t, gatewayapiMetrics)
	testUDPRoute(t, gatewayapiMetrics)
	testTLSRoute(t, gatewayapiMetrics)
	testBackendTLSPolicy(t, gatewayapiMetrics)
	testRateLimitPolicy(t, gatewayapiMetrics)
	testTLSPolicy(t, gatewayapiMetrics)
	testAuthPolicy(t, gatewayapiMetrics)
	testDNSPolicy(t, gatewayapiMetrics)
}

func TestKuadrantMetricsAvailable(t *testing.T) {
	buf := &bytes.Buffer{}

	err := framework.KsmClient.Metrics(buf)
	if err != nil {
		t.Fatalf("failed to get metrics from kube-state-metrics: %v", err)
	}

	re := regexp.MustCompile(`^(kuadrant_.*){(.*)}\s+(.*)`)
	scanner := bufio.NewScanner(buf)
	kuadrantMetrics := map[string][][]string{}
	for scanner.Scan() {
		// fmt.Printf("checking metric text=%s\n", scanner.Text())
		params := re.FindStringSubmatch(scanner.Text())
		// fmt.Printf("params=%v\n", params)
		if len(params) < 4 {
			continue
		}
		if kuadrantMetrics[params[1]] == nil {
			kuadrantMetrics[params[1]] = [][]string{}
		}
		fmt.Printf("Adding matched metric params=%v\n", params)
		kuadrantMetrics[params[1]] = append(kuadrantMetrics[params[1]], params)
	}
	testDNSRecord(t, kuadrantMetrics)
}

func testGatewayClasses(t *testing.T, metrics map[string][][]string) {
	//gatewayapi_gatewayclass_info
	gatewayClassInfo := metrics["gatewayapi_gatewayclass_info"]
	gatewayClass1Info := gatewayClassInfo[0]
	expectEqual(t, gatewayClass1Info[3], "1", "gatewayapi_gatewayclass_info__1 value")
	gatewayClass1InfoLabels := parseLabels(string(gatewayClass1Info[2]))
	expectEqual(t, gatewayClass1InfoLabels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_gatewayclass_info__1 customresource_group")
	expectEqual(t, gatewayClass1InfoLabels["customresource_kind"], "GatewayClass", "gatewayapi_gatewayclass_info__1 customresource_kind")
	expectEqual(t, gatewayClass1InfoLabels["customresource_version"], "v1beta1", "gatewayapi_gatewayclass_info__1 customresource_version")
	expectEqual(t, gatewayClass1InfoLabels["name"], "testgatewayclass1", "gatewayapi_gatewayclass_info__1 name")
	expectEqual(t, gatewayClass1InfoLabels["controller_name"], "example.com/gateway-controller", "gatewayapi_gatewayclass_info__1 controller_name")

	//gatewayapi_gatewayclass_status
	gatewayClassStatus := metrics["gatewayapi_gatewayclass_status"]
	gatewayClass1Status1 := gatewayClassStatus[0]
	expectEqual(t, gatewayClass1Status1[3], "1", "gatewayapi_gatewayclass_status__1 value")
	gatewayClass1Status1Labels := parseLabels(string(gatewayClass1Status1[2]))
	expectEqual(t, gatewayClass1Status1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_gatewayclass_status__1 customresource_group")
	expectEqual(t, gatewayClass1Status1Labels["customresource_kind"], "GatewayClass", "gatewayapi_gatewayclass_status__1 customresource_kind")
	expectEqual(t, gatewayClass1Status1Labels["customresource_version"], "v1beta1", "gatewayapi_gatewayclass_status__1 customresource_version")
	expectEqual(t, gatewayClass1Status1Labels["name"], "testgatewayclass1", "gatewayapi_gatewayclass_status__1 name")
	expectEqual(t, gatewayClass1Status1Labels["type"], "Accepted", "gatewayapi_gatewayclass_status__1 type")

	//gatewayapi_gatewayclass_status_supported_features
	gatewayClassStatusSupportedFeatures := metrics["gatewayapi_gatewayclass_status_supported_features"]
	gatewayClass1StatusSupportedFeatures1 := gatewayClassStatusSupportedFeatures[0]
	expectEqual(t, gatewayClass1StatusSupportedFeatures1[3], "1", "gatewayapi_gatewayclass_status_supported_features__1 value")
	gatewayClass1StatusSupportedFeatures1Labels := parseLabels(string(gatewayClass1StatusSupportedFeatures1[2]))
	expectEqual(t, gatewayClass1StatusSupportedFeatures1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_gatewayclass_status_supported_features__1 customresource_group")
	expectEqual(t, gatewayClass1StatusSupportedFeatures1Labels["customresource_kind"], "GatewayClass", "gatewayapi_gatewayclass_status_supported_features__1 customresource_kind")
	expectEqual(t, gatewayClass1StatusSupportedFeatures1Labels["customresource_version"], "v1beta1", "gatewayapi_gatewayclass_status_supported_features__1 customresource_version")
	expectEqual(t, gatewayClass1StatusSupportedFeatures1Labels["name"], "testgatewayclass1", "gatewayapi_gatewayclass_status_supported_features__1 name")

	expectedFeatures := map[int]string{
		0: "HTTPRoute",
		1: "HTTPRouteHostRewrite",
		2: "HTTPRoutePortRedirect",
		3: "HTTPRouteQueryParamMatching",
	}

	for i, feature := range gatewayClassStatusSupportedFeatures {
		featureInfo := parseLabels(string(feature[0]))
		featureName := featureInfo["features"]
		expectEqual(t, featureName, expectedFeatures[i], "gatewayapi_gatewayclass_status_supported_features__"+strconv.Itoa(i)+" features")
	}
}

func testGateways(t *testing.T, metrics map[string][][]string) {
	// gatewayapi_gateway_info
	gatewayInfo := metrics["gatewayapi_gateway_info"]
	gateway1Info := gatewayInfo[0]
	expectEqual(t, gateway1Info[3], "1", "gatewayapi_gateway_info__1 value")
	gateway1InfoLabels := parseLabels(string(gateway1Info[2]))
	expectEqual(t, gateway1InfoLabels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_gateway_info__1 customresource_group")
	expectEqual(t, gateway1InfoLabels["customresource_kind"], "Gateway", "gatewayapi_gateway_info__1 customresource_kind")
	expectEqual(t, gateway1InfoLabels["customresource_version"], "v1beta1", "gatewayapi_gateway_info__1 customresource_version")
	expectEqual(t, gateway1InfoLabels["name"], "testgateway1", "gatewayapi_gateway_info__1 name")
	expectEqual(t, gateway1InfoLabels["namespace"], "default", "gatewayapi_gateway_info__1 namespace")
	expectEqual(t, gateway1InfoLabels["gatewayclass_name"], "testgatewayclass1", "gatewayapi_gateway_info__1 gatewayclass_name")

	// gatewayapi_gateway_created
	gatewayCreated := metrics["gatewayapi_gateway_created"]
	gateway1Created := gatewayCreated[0]
	expectValidTimestampInPast(t, gateway1Created[3], "gatewayapi_gateway_created__1 value")
	gateway1CreatedLabels := parseLabels(string(gateway1Created[2]))
	expectEqual(t, gateway1CreatedLabels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_gateway_created__1 customresource_group")
	expectEqual(t, gateway1CreatedLabels["customresource_kind"], "Gateway", "gatewayapi_gateway_created__1 customresource_kind")
	expectEqual(t, gateway1CreatedLabels["customresource_version"], "v1beta1", "gatewayapi_gateway_created__1 customresource_version")
	expectEqual(t, gateway1CreatedLabels["name"], "testgateway1", "gatewayapi_gateway_created__1 name")
	expectEqual(t, gateway1CreatedLabels["namespace"], "default", "gatewayapi_gateway_created__1 namespace")

	//gatewayapi_gateway_listener_info
	gatewayListenerInfo := metrics["gatewayapi_gateway_listener_info"]
	gateway1ListenerInfo := gatewayListenerInfo[0]
	expectEqual(t, gateway1ListenerInfo[3], "1", "gatewayapi_gateway_listener_info__1 value")
	gateway1ListenerInfoLabels := parseLabels(string(gateway1ListenerInfo[2]))
	expectEqual(t, gateway1ListenerInfoLabels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_gateway_listener_info__1 customresource_group")
	expectEqual(t, gateway1ListenerInfoLabels["customresource_kind"], "Gateway", "gatewayapi_gateway_listener_info__1 customresource_kind")
	expectEqual(t, gateway1ListenerInfoLabels["customresource_version"], "v1beta1", "gatewayapi_gateway_listener_info__1 customresource_version")
	expectEqual(t, gateway1ListenerInfoLabels["name"], "testgateway1", "gatewayapi_gateway_listener_info__1 name")
	expectEqual(t, gateway1ListenerInfoLabels["namespace"], "default", "gatewayapi_gateway_listener_info__1 namespace")
	expectEqual(t, gateway1ListenerInfoLabels["listener_name"], "http", "gatewayapi_gateway_listener_info__1 listener name")
	expectEqual(t, gateway1ListenerInfoLabels["port"], "80", "gatewayapi_gateway_listener_info__1 port")
	expectEqual(t, gateway1ListenerInfoLabels["protocol"], "HTTP", "gatewayapi_gateway_listener_info__1 protocol")

	//gatewayapi_gateway_status
	gatewayStatus := metrics["gatewayapi_gateway_status"]
	gateway1Status1 := gatewayStatus[0]
	expectEqual(t, gateway1Status1[3], "1", "gatewayapi_gateway_status__1 value")
	gateway1Status1Labels := parseLabels(string(gateway1Status1[2]))
	expectEqual(t, gateway1Status1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_gateway_status__1 customresource_group")
	expectEqual(t, gateway1Status1Labels["customresource_kind"], "Gateway", "gatewayapi_gateway_status__1 customresource_kind")
	expectEqual(t, gateway1Status1Labels["customresource_version"], "v1beta1", "gatewayapi_gateway_status__1 customresource_version")
	expectEqual(t, gateway1Status1Labels["name"], "testgateway1", "gatewayapi_gateway_status__1 name")
	expectEqual(t, gateway1Status1Labels["namespace"], "default", "gatewayapi_gateway_status__1 namespace")
	expectEqual(t, gateway1Status1Labels["type"], "Accepted", "gatewayapi_gateway_status__1 type")
	gateway1Status2 := gatewayStatus[1]
	expectEqual(t, gateway1Status2[3], "1", "gatewayapi_gateway_status__2 value")
	gateway1Status2Labels := parseLabels(string(gateway1Status2[2]))
	expectEqual(t, gateway1Status2Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_gateway_status__2 customresource_group")
	expectEqual(t, gateway1Status2Labels["customresource_kind"], "Gateway", "gatewayapi_gateway_status__2 customresource_kind")
	expectEqual(t, gateway1Status2Labels["customresource_version"], "v1beta1", "gatewayapi_gateway_status__2 customresource_version")
	expectEqual(t, gateway1Status2Labels["name"], "testgateway1", "gatewayapi_gateway_status__2 name")
	expectEqual(t, gateway1Status2Labels["namespace"], "default", "gatewayapi_gateway_status__2 namespace")
	expectEqual(t, gateway1Status2Labels["type"], "Programmed", "gatewayapi_gateway_status__2 type")

	//gatewayapi_gateway_status_listener_attached_routes
	gatewayStatusListenerAttachedRoutes := metrics["gatewayapi_gateway_status_listener_attached_routes"]
	gateway1StatusListenerAttachedRoutes1 := gatewayStatusListenerAttachedRoutes[0]
	expectEqual(t, gateway1StatusListenerAttachedRoutes1[3], "2", "gatewayapi_gateway_status_listener_attached_routes__1 value")
	gateway1StatusListenerAttachedRoutes1Labels := parseLabels(string(gateway1StatusListenerAttachedRoutes1[2]))
	expectEqual(t, gateway1StatusListenerAttachedRoutes1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_gateway_status_listener_attached_routes__1 customresource_group")
	expectEqual(t, gateway1StatusListenerAttachedRoutes1Labels["customresource_kind"], "Gateway", "gatewayapi_gateway_status_listener_attached_routes__1 customresource_kind")
	expectEqual(t, gateway1StatusListenerAttachedRoutes1Labels["customresource_version"], "v1beta1", "gatewayapi_gateway_status_listener_attached_routes__1 customresource_version")
	expectEqual(t, gateway1StatusListenerAttachedRoutes1Labels["name"], "testgateway1", "gatewayapi_gateway_status_listener_attached_routes__1 name")
	expectEqual(t, gateway1StatusListenerAttachedRoutes1Labels["namespace"], "default", "gatewayapi_gateway_status_listener_attached_routes__1 namespace")
	expectEqual(t, gateway1StatusListenerAttachedRoutes1Labels["listener_name"], "http", "gatewayapi_gateway_status_listener_attached_routes__1 listener_name")

	//gatewayapi_gateway_status_address_info
	gatewayStatusAddressInfo := metrics["gatewayapi_gateway_status_address_info"]
	gateway1StatusAddressInfo1 := gatewayStatusAddressInfo[0]
	expectEqual(t, gateway1StatusAddressInfo1[3], "1", "gatewayapi_gateway_status_address_info__1 value")
	gateway1StatusAddressInfo1Labels := parseLabels(string(gateway1StatusAddressInfo1[2]))
	expectEqual(t, gateway1StatusAddressInfo1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_gateway_status_address_info__1 customresource_group")
	expectEqual(t, gateway1StatusAddressInfo1Labels["customresource_kind"], "Gateway", "gatewayapi_gateway_status_address_info__1 customresource_kind")
	expectEqual(t, gateway1StatusAddressInfo1Labels["customresource_version"], "v1beta1", "gatewayapi_gateway_status_address_info__1 customresource_version")
	expectEqual(t, gateway1StatusAddressInfo1Labels["name"], "testgateway1", "gatewayapi_gateway_status_address_info__1 name")
	expectEqual(t, gateway1StatusAddressInfo1Labels["namespace"], "default", "gatewayapi_gateway_status_address_info__1 namespace")
	expectEqual(t, gateway1StatusAddressInfo1Labels["type"], "Hostname", "gatewayapi_gateway_status_address_info__1 type")
	expectEqual(t, gateway1StatusAddressInfo1Labels["value"], "localhost", "gatewayapi_gateway_status_address_info__1 value")
	gateway1StatusAddressInfo2 := gatewayStatusAddressInfo[1]
	expectEqual(t, gateway1StatusAddressInfo2[3], "1", "gatewayapi_gateway_status_address_info__2 value")
	gateway1StatusAddressInfo2Labels := parseLabels(string(gateway1StatusAddressInfo2[2]))
	expectEqual(t, gateway1StatusAddressInfo2Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_gateway_status_address_info__2 customresource_group")
	expectEqual(t, gateway1StatusAddressInfo2Labels["customresource_kind"], "Gateway", "gatewayapi_gateway_status_address_info__2 customresource_kind")
	expectEqual(t, gateway1StatusAddressInfo2Labels["customresource_version"], "v1beta1", "gatewayapi_gateway_status_address_info__2 customresource_version")
	expectEqual(t, gateway1StatusAddressInfo2Labels["name"], "testgateway1", "gatewayapi_gateway_status_address_info__2 name")
	expectEqual(t, gateway1StatusAddressInfo2Labels["namespace"], "default", "gatewayapi_gateway_status_address_info__2 namespace")
	expectEqual(t, gateway1StatusAddressInfo2Labels["type"], "IPAddress", "gatewayapi_gateway_status_address_info__2 type")
	expectEqual(t, gateway1StatusAddressInfo2Labels["value"], "127.0.0.1", "gatewayapi_gateway_status_address_info__2 value")
}

func testHTTPRoutes(t *testing.T, metrics map[string][][]string) {
	// gatewayapi_httproute_created
	httprouteCreated := metrics["gatewayapi_httproute_created"]
	httproute1Created := httprouteCreated[0]
	expectValidTimestampInPast(t, httproute1Created[3], "gatewayapi_httproute_created__1 value")
	httproute1CreatedLabels := parseLabels(string(httproute1Created[2]))
	expectEqual(t, httproute1CreatedLabels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_httproute_created__1 customresource_group")
	expectEqual(t, httproute1CreatedLabels["customresource_kind"], "HTTPRoute", "gatewayapi_httproute_created__1 customresource_kind")
	expectEqual(t, httproute1CreatedLabels["customresource_version"], "v1beta1", "gatewayapi_httproute_created__1 customresource_version")
	expectEqual(t, httproute1CreatedLabels["name"], "testroute1", "gatewayapi_httproute_created__1 name")
	expectEqual(t, httproute1CreatedLabels["namespace"], "default", "gatewayapi_httproute_created__1 namespace")

	//gatewayapi_httproute_hostname_info
	httprouteHostnameInfo := metrics["gatewayapi_httproute_hostname_info"]
	httproute1HostnameInfo1 := httprouteHostnameInfo[0]
	expectEqual(t, httproute1HostnameInfo1[3], "1", "gatewayapi_httproute_hostname_info__1 value")
	httproute1HostnameInfo1Labels := parseLabels(string(httproute1HostnameInfo1[2]))
	expectEqual(t, httproute1HostnameInfo1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_httproute_hostname_info__1 customresource_group")
	expectEqual(t, httproute1HostnameInfo1Labels["customresource_kind"], "HTTPRoute", "gatewayapi_httproute_hostname_info__1 customresource_kind")
	expectEqual(t, httproute1HostnameInfo1Labels["customresource_version"], "v1beta1", "gatewayapi_httproute_hostname_info__1 customresource_version")
	expectEqual(t, httproute1HostnameInfo1Labels["name"], "testroute1", "gatewayapi_httproute_hostname_info__1 name")
	expectEqual(t, httproute1HostnameInfo1Labels["namespace"], "default", "gatewayapi_httproute_hostname_info__1 namespace")
	expectEqual(t, httproute1HostnameInfo1Labels["hostname"], "test1.example.com", "gatewayapi_httproute_hostname_info__1 hostname")

	//gatewayapi_httproute_parent_info
	httprouteParentInfo := metrics["gatewayapi_httproute_parent_info"]
	httproute1ParentInfo1 := httprouteParentInfo[0]
	expectEqual(t, httproute1ParentInfo1[3], "1", "gatewayapi_httproute_parent_info__1 value")
	httproute1ParentInfo1Labels := parseLabels(string(httproute1ParentInfo1[2]))
	expectEqual(t, httproute1ParentInfo1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_httproute_parent_info__1 customresource_group")
	expectEqual(t, httproute1ParentInfo1Labels["customresource_kind"], "HTTPRoute", "gatewayapi_httproute_parent_info__1 customresource_kind")
	expectEqual(t, httproute1ParentInfo1Labels["customresource_version"], "v1beta1", "gatewayapi_httproute_parent_info__1 customresource_version")
	expectEqual(t, httproute1ParentInfo1Labels["name"], "testroute1", "gatewayapi_httproute_parent_info__1 name")
	expectEqual(t, httproute1ParentInfo1Labels["namespace"], "default", "gatewayapi_httproute_parent_info__1 namespace")
	expectEqual(t, httproute1ParentInfo1Labels["parent_group"], "gateway.networking.k8s.io", "gatewayapi_httproute_parent_info__1 parent_group")
	expectEqual(t, httproute1ParentInfo1Labels["parent_kind"], "Gateway", "gatewayapi_httproute_parent_info__1 parent_kind")
	expectEqual(t, httproute1ParentInfo1Labels["parent_namespace"], "default", "gatewayapi_httproute_parent_info__1 parent_namespace")
	expectEqual(t, httproute1ParentInfo1Labels["parent_name"], "testgateway1", "gatewayapi_httproute_parent_info__1 parent_name")

	//gatewayapi_httproute_status_parent_info
	httprouteParentStatusInfo := metrics["gatewayapi_httproute_status_parent_info"]
	httproute1ParentStatusInfo1 := httprouteParentStatusInfo[0]
	expectEqual(t, httproute1ParentStatusInfo1[3], "1", "gatewayapi_httproute_status_parent_info__1 value")
	httproute1ParentStatusInfo1Labels := parseLabels(string(httproute1ParentStatusInfo1[2]))
	expectEqual(t, httproute1ParentStatusInfo1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_httproute_status_parent_info__1 customresource_group")
	expectEqual(t, httproute1ParentStatusInfo1Labels["customresource_kind"], "HTTPRoute", "gatewayapi_httproute_status_parent_info__1 customresource_kind")
	expectEqual(t, httproute1ParentStatusInfo1Labels["customresource_version"], "v1beta1", "gatewayapi_httproute_status_parent_info__1 customresource_version")
	expectEqual(t, httproute1ParentStatusInfo1Labels["name"], "testroute1", "gatewayapi_httproute_status_parent_info__1 name")
	expectEqual(t, httproute1ParentStatusInfo1Labels["namespace"], "default", "gatewayapi_httproute_status_parent_info__1 namespace")
	expectEqual(t, httproute1ParentStatusInfo1Labels["parent_group"], "gateway.networking.k8s.io", "gatewayapi_httproute_status_parent_info__1 parent_group")
	expectEqual(t, httproute1ParentStatusInfo1Labels["parent_kind"], "Gateway", "gatewayapi_httproute_status_parent_info__1 parent_kind")
	expectEqual(t, httproute1ParentStatusInfo1Labels["parent_namespace"], "default", "gatewayapi_httproute_status_parent_info__1 parent_namespace")
	expectEqual(t, httproute1ParentStatusInfo1Labels["parent_name"], "testgateway1", "gatewayapi_httproute_status_parent_info__1 parent_name")
}

func testGRPCRoutes(t *testing.T, metrics map[string][][]string) {
	// gatewayapi_grpcroute_created
	grpcrouteCreated := metrics["gatewayapi_grpcroute_created"]
	grpcroute1Created := grpcrouteCreated[0]
	expectValidTimestampInPast(t, grpcroute1Created[3], "gatewayapi_grpcroute_created__1 value")
	grpcroute1CreatedLabels := parseLabels(string(grpcroute1Created[2]))
	expectEqual(t, grpcroute1CreatedLabels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_grpcroute_created__1 customresource_group")
	expectEqual(t, grpcroute1CreatedLabels["customresource_kind"], "GRPCRoute", "gatewayapi_grpcroute_created__1 customresource_kind")
	expectEqual(t, grpcroute1CreatedLabels["customresource_version"], "v1alpha2", "gatewayapi_grpcroute_created__1 customresource_version")
	expectEqual(t, grpcroute1CreatedLabels["name"], "testgrpcroute1", "gatewayapi_grpcroute_created__1 name")
	expectEqual(t, grpcroute1CreatedLabels["namespace"], "default", "gatewayapi_grpcroute_created__1 namespace")

	//gatewayapi_grpcroute_hostname_info
	grpcrouteHostnameInfo := metrics["gatewayapi_grpcroute_hostname_info"]
	grpcroute1HostnameInfo1 := grpcrouteHostnameInfo[0]
	expectEqual(t, grpcroute1HostnameInfo1[3], "1", "gatewayapi_grpcroute_hostname_info__1 value")
	grpcroute1HostnameInfo1Labels := parseLabels(string(grpcroute1HostnameInfo1[2]))
	expectEqual(t, grpcroute1HostnameInfo1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_grpcroute_hostname_info__1 customresource_group")
	expectEqual(t, grpcroute1HostnameInfo1Labels["customresource_kind"], "GRPCRoute", "gatewayapi_grpcroute_hostname_info__1 customresource_kind")
	expectEqual(t, grpcroute1HostnameInfo1Labels["customresource_version"], "v1alpha2", "gatewayapi_grpcroute_hostname_info__1 customresource_version")
	expectEqual(t, grpcroute1HostnameInfo1Labels["name"], "testgrpcroute1", "gatewayapi_grpcroute_hostname_info__1 name")
	expectEqual(t, grpcroute1HostnameInfo1Labels["namespace"], "default", "gatewayapi_grpcroute_hostname_info__1 namespace")
	expectEqual(t, grpcroute1HostnameInfo1Labels["hostname"], "test1.example.com", "gatewayapi_grpcroute_hostname_info__1 hostname")

	//gatewayapi_grpcroute_parent_info
	grpcrouteParentInfo := metrics["gatewayapi_grpcroute_parent_info"]
	grpcroute1ParentInfo1 := grpcrouteParentInfo[0]
	expectEqual(t, grpcroute1ParentInfo1[3], "1", "gatewayapi_grpcroute_parent_info__1 value")
	grpcroute1ParentInfo1Labels := parseLabels(string(grpcroute1ParentInfo1[2]))
	expectEqual(t, grpcroute1ParentInfo1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_grpcroute_parent_info__1 customresource_group")
	expectEqual(t, grpcroute1ParentInfo1Labels["customresource_kind"], "GRPCRoute", "gatewayapi_grpcroute_parent_info__1 customresource_kind")
	expectEqual(t, grpcroute1ParentInfo1Labels["customresource_version"], "v1alpha2", "gatewayapi_grpcroute_parent_info__1 customresource_version")
	expectEqual(t, grpcroute1ParentInfo1Labels["name"], "testgrpcroute1", "gatewayapi_grpcroute_parent_info__1 name")
	expectEqual(t, grpcroute1ParentInfo1Labels["namespace"], "default", "gatewayapi_grpcroute_parent_info__1 namespace")
	expectEqual(t, grpcroute1ParentInfo1Labels["parent_group"], "gateway.networking.k8s.io", "gatewayapi_grpcroute_parent_info__1 parent_group")
	expectEqual(t, grpcroute1ParentInfo1Labels["parent_kind"], "Gateway", "gatewayapi_grpcroute_parent_info__1 parent_kind")
	expectEqual(t, grpcroute1ParentInfo1Labels["parent_namespace"], "default", "gatewayapi_grpcroute_parent_info__1 parent_namespace")
	expectEqual(t, grpcroute1ParentInfo1Labels["parent_name"], "testgateway1", "gatewayapi_grpcroute_parent_info__1 parent_name")

	//gatewayapi_grpcroute_status_parent_info
	grpcrouteParentStatusInfo := metrics["gatewayapi_grpcroute_status_parent_info"]
	grpcroute1ParentStatusInfo1 := grpcrouteParentStatusInfo[0]
	expectEqual(t, grpcroute1ParentStatusInfo1[3], "1", "gatewayapi_grpcroute_status_parent_info__1 value")
	grpcroute1ParentStatusInfo1Labels := parseLabels(string(grpcroute1ParentStatusInfo1[2]))
	expectEqual(t, grpcroute1ParentStatusInfo1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_grpcroute_status_parent_info__1 customresource_group")
	expectEqual(t, grpcroute1ParentStatusInfo1Labels["customresource_kind"], "GRPCRoute", "gatewayapi_grpcroute_status_parent_info__1 customresource_kind")
	expectEqual(t, grpcroute1ParentStatusInfo1Labels["customresource_version"], "v1alpha2", "gatewayapi_grpcroute_status_parent_info__1 customresource_version")
	expectEqual(t, grpcroute1ParentStatusInfo1Labels["name"], "testgrpcroute1", "gatewayapi_grpcroute_status_parent_info__1 name")
	expectEqual(t, grpcroute1ParentStatusInfo1Labels["namespace"], "default", "gatewayapi_grpcroute_status_parent_info__1 namespace")
	expectEqual(t, grpcroute1ParentStatusInfo1Labels["parent_group"], "gateway.networking.k8s.io", "gatewayapi_grpcroute_status_parent_info__1 parent_group")
	expectEqual(t, grpcroute1ParentStatusInfo1Labels["parent_kind"], "Gateway", "gatewayapi_grpcroute_status_parent_info__1 parent_kind")
	expectEqual(t, grpcroute1ParentStatusInfo1Labels["parent_namespace"], "default", "gatewayapi_grpcroute_status_parent_info__1 parent_namespace")
	expectEqual(t, grpcroute1ParentStatusInfo1Labels["parent_name"], "testgateway1", "gatewayapi_grpcroute_status_parent_info__1 parent_name")
}

func testTLSRoute(t *testing.T, metrics map[string][][]string) {
	// gatewayapi_tlsroute_created
	tlsrouteCreated := metrics["gatewayapi_tlsroute_created"]
	tlsroute1Created := tlsrouteCreated[0]
	expectValidTimestampInPast(t, tlsroute1Created[3], "gatewayapi_tlsroute_created__1 value")
	tlsroute1CreatedLabels := parseLabels(string(tlsroute1Created[2]))
	expectEqual(t, tlsroute1CreatedLabels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_tlsroute_created__1 customresource_group")
	expectEqual(t, tlsroute1CreatedLabels["customresource_kind"], "TLSRoute", "gatewayapi_tlsroute_created__1 customresource_kind")
	expectEqual(t, tlsroute1CreatedLabels["customresource_version"], "v1alpha2", "gatewayapi_tlsroute_created__1 customresource_version")
	expectEqual(t, tlsroute1CreatedLabels["name"], "testtlsroute1", "gatewayapi_tlsroute_created__1 name")
	expectEqual(t, tlsroute1CreatedLabels["namespace"], "default", "gatewayapi_tlsroute_created__1 namespace")

	//gatewayapi_tlsroute_hostname_info
	tlsrouteHostnameInfo := metrics["gatewayapi_tlsroute_hostname_info"]
	tlsroute1HostnameInfo1 := tlsrouteHostnameInfo[0]
	expectEqual(t, tlsroute1HostnameInfo1[3], "1", "gatewayapi_tlsroute_hostname_info__1 value")
	tlsroute1HostnameInfo1Labels := parseLabels(string(tlsroute1HostnameInfo1[2]))
	expectEqual(t, tlsroute1HostnameInfo1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_tlsroute_hostname_info__1 customresource_group")
	expectEqual(t, tlsroute1HostnameInfo1Labels["customresource_kind"], "TLSRoute", "gatewayapi_tlsroute_hostname_info__1 customresource_kind")
	expectEqual(t, tlsroute1HostnameInfo1Labels["customresource_version"], "v1alpha2", "gatewayapi_tlsroute_hostname_info__1 customresource_version")
	expectEqual(t, tlsroute1HostnameInfo1Labels["name"], "testtlsroute1", "gatewayapi_tlsroute_hostname_info__1 name")
	expectEqual(t, tlsroute1HostnameInfo1Labels["namespace"], "default", "gatewayapi_tlsroute_hostname_info__1 namespace")
	expectEqual(t, tlsroute1HostnameInfo1Labels["hostname"], "test1.example.com", "gatewayapi_tlsroute_hostname_info__1 hostname")

	//gatewayapi_tlsroute_parent_info
	tlsrouteParentInfo := metrics["gatewayapi_tlsroute_parent_info"]
	tlsroute1ParentInfo1 := tlsrouteParentInfo[0]
	expectEqual(t, tlsroute1ParentInfo1[3], "1", "gatewayapi_tlsroute_parent_info__1 value")
	tlsroute1ParentInfo1Labels := parseLabels(string(tlsroute1ParentInfo1[2]))
	expectEqual(t, tlsroute1ParentInfo1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_tlsroute_parent_info__1 customresource_group")
	expectEqual(t, tlsroute1ParentInfo1Labels["customresource_kind"], "TLSRoute", "gatewayapi_tlsroute_parent_info__1 customresource_kind")
	expectEqual(t, tlsroute1ParentInfo1Labels["customresource_version"], "v1alpha2", "gatewayapi_tlsroute_parent_info__1 customresource_version")
	expectEqual(t, tlsroute1ParentInfo1Labels["name"], "testtlsroute1", "gatewayapi_tlsroute_parent_info__1 name")
	expectEqual(t, tlsroute1ParentInfo1Labels["namespace"], "default", "gatewayapi_tlsroute_parent_info__1 namespace")
	expectEqual(t, tlsroute1ParentInfo1Labels["parent_group"], "gateway.networking.k8s.io", "gatewayapi_tlsroute_parent_info__1 parent_group")
	expectEqual(t, tlsroute1ParentInfo1Labels["parent_kind"], "Gateway", "gatewayapi_tlsroute_parent_info__1 parent_kind")
	expectEqual(t, tlsroute1ParentInfo1Labels["parent_namespace"], "default", "gatewayapi_tlsroute_parent_info__1 parent_namespace")
	expectEqual(t, tlsroute1ParentInfo1Labels["parent_name"], "testgateway1", "gatewayapi_tlsroute_parent_info__1 parent_name")

	//gatewayapi_tlsroute_status_parent_info
	tlsrouteParentStatusInfo := metrics["gatewayapi_tlsroute_status_parent_info"]
	tlsroute1ParentStatusInfo1 := tlsrouteParentStatusInfo[0]
	expectEqual(t, tlsroute1ParentStatusInfo1[3], "1", "gatewayapi_tlsroute_status_parent_info__1 value")
	tlsroute1ParentStatusInfo1Labels := parseLabels(string(tlsroute1ParentStatusInfo1[2]))
	expectEqual(t, tlsroute1ParentStatusInfo1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_tlsroute_status_parent_info__1 customresource_group")
	expectEqual(t, tlsroute1ParentStatusInfo1Labels["customresource_kind"], "TLSRoute", "gatewayapi_tlsroute_status_parent_info__1 customresource_kind")
	expectEqual(t, tlsroute1ParentStatusInfo1Labels["customresource_version"], "v1alpha2", "gatewayapi_tlsroute_status_parent_info__1 customresource_version")
	expectEqual(t, tlsroute1ParentStatusInfo1Labels["name"], "testtlsroute1", "gatewayapi_tlsroute_status_parent_info__1 name")
	expectEqual(t, tlsroute1ParentStatusInfo1Labels["namespace"], "default", "gatewayapi_tlsroute_status_parent_info__1 namespace")
	expectEqual(t, tlsroute1ParentStatusInfo1Labels["parent_group"], "gateway.networking.k8s.io", "gatewayapi_tlsroute_status_parent_info__1 parent_group")
	expectEqual(t, tlsroute1ParentStatusInfo1Labels["parent_kind"], "Gateway", "gatewayapi_tlsroute_status_parent_info__1 parent_kind")
	expectEqual(t, tlsroute1ParentStatusInfo1Labels["parent_namespace"], "default", "gatewayapi_tlsroute_status_parent_info__1 parent_namespace")
	expectEqual(t, tlsroute1ParentStatusInfo1Labels["parent_name"], "testgateway1", "gatewayapi_tlsroute_status_parent_info__1 parent_name")
}

func testTCPRoute(t *testing.T, metrics map[string][][]string) {
	// gatewayapi_tcproute_created
	tcprouteCreated := metrics["gatewayapi_tcproute_created"]
	tcproute1Created := tcprouteCreated[0]
	expectValidTimestampInPast(t, tcproute1Created[3], "gatewayapi_tcproute_created__1 value")
	tcproute1CreatedLabels := parseLabels(string(tcproute1Created[2]))
	expectEqual(t, tcproute1CreatedLabels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_tcproute_created__1 customresource_group")
	expectEqual(t, tcproute1CreatedLabels["customresource_kind"], "TCPRoute", "gatewayapi_tcproute_created__1 customresource_kind")
	expectEqual(t, tcproute1CreatedLabels["customresource_version"], "v1alpha2", "gatewayapi_tcproute_created__1 customresource_version")
	expectEqual(t, tcproute1CreatedLabels["name"], "testtcproute1", "gatewayapi_tcproute_created__1 name")
	expectEqual(t, tcproute1CreatedLabels["namespace"], "default", "gatewayapi_tcproute_created__1 namespace")

	//gatewayapi_tcproute_parent_info
	tcprouteParentInfo := metrics["gatewayapi_tcproute_parent_info"]
	tcproute1ParentInfo1 := tcprouteParentInfo[0]
	expectEqual(t, tcproute1ParentInfo1[3], "1", "gatewayapi_tcproute_parent_info__1 value")
	tcproute1ParentInfo1Labels := parseLabels(string(tcproute1ParentInfo1[2]))
	expectEqual(t, tcproute1ParentInfo1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_tcproute_parent_info__1 customresource_group")
	expectEqual(t, tcproute1ParentInfo1Labels["customresource_kind"], "TCPRoute", "gatewayapi_tcproute_parent_info__1 customresource_kind")
	expectEqual(t, tcproute1ParentInfo1Labels["customresource_version"], "v1alpha2", "gatewayapi_tcproute_parent_info__1 customresource_version")
	expectEqual(t, tcproute1ParentInfo1Labels["name"], "testtcproute1", "gatewayapi_tcproute_parent_info__1 name")
	expectEqual(t, tcproute1ParentInfo1Labels["namespace"], "default", "gatewayapi_tcproute_parent_info__1 namespace")
	expectEqual(t, tcproute1ParentInfo1Labels["parent_group"], "gateway.networking.k8s.io", "gatewayapi_tcproute_parent_info__1 parent_group")
	expectEqual(t, tcproute1ParentInfo1Labels["parent_kind"], "Gateway", "gatewayapi_tcproute_parent_info__1 parent_kind")
	expectEqual(t, tcproute1ParentInfo1Labels["parent_namespace"], "default", "gatewayapi_tcproute_parent_info__1 parent_namespace")
	expectEqual(t, tcproute1ParentInfo1Labels["parent_name"], "testgateway1", "gatewayapi_tcproute_parent_info__1 parent_name")

	//gatewayapi_tcproute_status_parent_info
	tcprouteParentStatusInfo := metrics["gatewayapi_tcproute_status_parent_info"]
	tcproute1ParentStatusInfo1 := tcprouteParentStatusInfo[0]
	expectEqual(t, tcproute1ParentStatusInfo1[3], "1", "gatewayapi_tcproute_status_parent_info__1 value")
	tcproute1ParentStatusInfo1Labels := parseLabels(string(tcproute1ParentStatusInfo1[2]))
	expectEqual(t, tcproute1ParentStatusInfo1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_tcproute_status_parent_info__1 customresource_group")
	expectEqual(t, tcproute1ParentStatusInfo1Labels["customresource_kind"], "TCPRoute", "gatewayapi_tcproute_status_parent_info__1 customresource_kind")
	expectEqual(t, tcproute1ParentStatusInfo1Labels["customresource_version"], "v1alpha2", "gatewayapi_tcproute_status_parent_info__1 customresource_version")
	expectEqual(t, tcproute1ParentStatusInfo1Labels["name"], "testtcproute1", "gatewayapi_tcproute_status_parent_info__1 name")
	expectEqual(t, tcproute1ParentStatusInfo1Labels["namespace"], "default", "gatewayapi_tcproute_status_parent_info__1 namespace")
	expectEqual(t, tcproute1ParentStatusInfo1Labels["parent_group"], "gateway.networking.k8s.io", "gatewayapi_tcproute_status_parent_info__1 parent_group")
	expectEqual(t, tcproute1ParentStatusInfo1Labels["parent_kind"], "Gateway", "gatewayapi_tcproute_status_parent_info__1 parent_kind")
	expectEqual(t, tcproute1ParentStatusInfo1Labels["parent_namespace"], "default", "gatewayapi_tcproute_status_parent_info__1 parent_namespace")
	expectEqual(t, tcproute1ParentStatusInfo1Labels["parent_name"], "testgateway1", "gatewayapi_tcproute_status_parent_info__1 parent_name")
}

func testUDPRoute(t *testing.T, metrics map[string][][]string) {
	// gatewayapi_udproute_created
	udprouteCreated := metrics["gatewayapi_udproute_created"]
	udproute1Created := udprouteCreated[0]
	expectValidTimestampInPast(t, udproute1Created[3], "gatewayapi_udproute_created__1 value")
	udproute1CreatedLabels := parseLabels(string(udproute1Created[2]))
	expectEqual(t, udproute1CreatedLabels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_udproute_created__1 customresource_group")
	expectEqual(t, udproute1CreatedLabels["customresource_kind"], "UDPRoute", "gatewayapi_udproute_created__1 customresource_kind")
	expectEqual(t, udproute1CreatedLabels["customresource_version"], "v1alpha2", "gatewayapi_udproute_created__1 customresource_version")
	expectEqual(t, udproute1CreatedLabels["name"], "testudproute1", "gatewayapi_udproute_created__1 name")
	expectEqual(t, udproute1CreatedLabels["namespace"], "default", "gatewayapi_udproute_created__1 namespace")

	//gatewayapi_udproute_parent_info
	udprouteParentInfo := metrics["gatewayapi_udproute_parent_info"]
	udproute1ParentInfo1 := udprouteParentInfo[0]
	expectEqual(t, udproute1ParentInfo1[3], "1", "gatewayapi_udproute_parent_info__1 value")
	udproute1ParentInfo1Labels := parseLabels(string(udproute1ParentInfo1[2]))
	expectEqual(t, udproute1ParentInfo1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_udproute_parent_info__1 customresource_group")
	expectEqual(t, udproute1ParentInfo1Labels["customresource_kind"], "UDPRoute", "gatewayapi_udproute_parent_info__1 customresource_kind")
	expectEqual(t, udproute1ParentInfo1Labels["customresource_version"], "v1alpha2", "gatewayapi_udproute_parent_info__1 customresource_version")
	expectEqual(t, udproute1ParentInfo1Labels["name"], "testudproute1", "gatewayapi_udproute_parent_info__1 name")
	expectEqual(t, udproute1ParentInfo1Labels["namespace"], "default", "gatewayapi_udproute_parent_info__1 namespace")
	expectEqual(t, udproute1ParentInfo1Labels["parent_group"], "gateway.networking.k8s.io", "gatewayapi_udproute_parent_info__1 parent_group")
	expectEqual(t, udproute1ParentInfo1Labels["parent_kind"], "Gateway", "gatewayapi_udproute_parent_info__1 parent_kind")
	expectEqual(t, udproute1ParentInfo1Labels["parent_namespace"], "default", "gatewayapi_udproute_parent_info__1 parent_namespace")
	expectEqual(t, udproute1ParentInfo1Labels["parent_name"], "testgateway1", "gatewayapi_udproute_parent_info__1 parent_name")

	//gatewayapi_udproute_status_parent_info
	udprouteParentStatusInfo := metrics["gatewayapi_udproute_status_parent_info"]
	udproute1ParentStatusInfo1 := udprouteParentStatusInfo[0]
	expectEqual(t, udproute1ParentStatusInfo1[3], "1", "gatewayapi_udproute_status_parent_info__1 value")
	udproute1ParentStatusInfo1Labels := parseLabels(string(udproute1ParentStatusInfo1[2]))
	expectEqual(t, udproute1ParentStatusInfo1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_udproute_status_parent_info__1 customresource_group")
	expectEqual(t, udproute1ParentStatusInfo1Labels["customresource_kind"], "UDPRoute", "gatewayapi_udproute_status_parent_info__1 customresource_kind")
	expectEqual(t, udproute1ParentStatusInfo1Labels["customresource_version"], "v1alpha2", "gatewayapi_udproute_status_parent_info__1 customresource_version")
	expectEqual(t, udproute1ParentStatusInfo1Labels["name"], "testudproute1", "gatewayapi_udproute_status_parent_info__1 name")
	expectEqual(t, udproute1ParentStatusInfo1Labels["namespace"], "default", "gatewayapi_udproute_status_parent_info__1 namespace")
	expectEqual(t, udproute1ParentStatusInfo1Labels["parent_group"], "gateway.networking.k8s.io", "gatewayapi_udproute_status_parent_info__1 parent_group")
	expectEqual(t, udproute1ParentStatusInfo1Labels["parent_kind"], "Gateway", "gatewayapi_udproute_status_parent_info__1 parent_kind")
	expectEqual(t, udproute1ParentStatusInfo1Labels["parent_namespace"], "default", "gatewayapi_udproute_status_parent_info__1 parent_namespace")
	expectEqual(t, udproute1ParentStatusInfo1Labels["parent_name"], "testgateway1", "gatewayapi_udproute_status_parent_info__1 parent_name")
}

func testBackendTLSPolicy(t *testing.T, metrics map[string][][]string) {
	// gatewayapi_backendtlspolicy_created
	backendtlspolicyCreated := metrics["gatewayapi_backendtlspolicy_created"]
	backendtlspolicy1Created := backendtlspolicyCreated[0]
	expectValidTimestampInPast(t, backendtlspolicy1Created[3], "gatewayapi_backendtlspolicy_created__1 value")
	backendtlspolicy1CreatedLabels := parseLabels(string(backendtlspolicy1Created[2]))
	expectEqual(t, backendtlspolicy1CreatedLabels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_backendtlspolicy_created__1 customresource_group")
	expectEqual(t, backendtlspolicy1CreatedLabels["customresource_kind"], "BackendTLSPolicy", "gatewayapi_backendtlspolicy_created__1 customresource_kind")
	expectEqual(t, backendtlspolicy1CreatedLabels["customresource_version"], "v1alpha2", "gatewayapi_backendtlspolicy_created__1 customresource_version")
	expectEqual(t, backendtlspolicy1CreatedLabels["name"], "testbackendtlspolicy1", "gatewayapi_backendtlspolicy_created__1 name")
	expectEqual(t, backendtlspolicy1CreatedLabels["namespace"], "default", "gatewayapi_backendtlspolicy_created__1 namespace")

	//gatewayapi_backendtlspolicy_target_info
	backendtlspolicyParentInfo := metrics["gatewayapi_backendtlspolicy_target_info"]
	backendtlspolicy1ParentInfo1 := backendtlspolicyParentInfo[0]
	expectEqual(t, backendtlspolicy1ParentInfo1[3], "1", "gatewayapi_backendtlspolicy_target_info__1 value")
	backendtlspolicy1ParentInfo1Labels := parseLabels(string(backendtlspolicy1ParentInfo1[2]))
	expectEqual(t, backendtlspolicy1ParentInfo1Labels["customresource_group"], "gateway.networking.k8s.io", "gatewayapi_backendtlspolicy_target_info__1 customresource_group")
	expectEqual(t, backendtlspolicy1ParentInfo1Labels["customresource_kind"], "BackendTLSPolicy", "gatewayapi_backendtlspolicy_target_info__1 customresource_kind")
	expectEqual(t, backendtlspolicy1ParentInfo1Labels["customresource_version"], "v1alpha2", "gatewayapi_backendtlspolicy_target_info__1 customresource_version")
	expectEqual(t, backendtlspolicy1ParentInfo1Labels["name"], "testbackendtlspolicy1", "gatewayapi_backendtlspolicy_target_info__1 name")
	expectEqual(t, backendtlspolicy1ParentInfo1Labels["namespace"], "default", "gatewayapi_backendtlspolicy_target_info__1 namespace")
	expectEqual(t, backendtlspolicy1ParentInfo1Labels["target_group"], "", "gatewayapi_backendtlspolicy_target_info__1 target_group")
	expectEqual(t, backendtlspolicy1ParentInfo1Labels["target_kind"], "Service", "gatewayapi_backendtlspolicy_target_info__1 target_kind")
}

func testRateLimitPolicy(t *testing.T, metrics map[string][][]string) {
	// gatewayapi_ratelimitpolicy_created
	ratelimitpolicyCreated := metrics["gatewayapi_ratelimitpolicy_created"]
	ratelimitpolicy1Created := ratelimitpolicyCreated[0]
	expectValidTimestampInPast(t, ratelimitpolicy1Created[3], "gatewayapi_ratelimitpolicy_created__1 value")
	ratelimitpolicy1CreatedLabels := parseLabels(string(ratelimitpolicy1Created[2]))
	expectEqual(t, ratelimitpolicy1CreatedLabels["customresource_group"], "kuadrant.io", "gatewayapi_ratelimitpolicy_created__1 customresource_group")
	expectEqual(t, ratelimitpolicy1CreatedLabels["customresource_kind"], "RateLimitPolicy", "gatewayapi_ratelimitpolicy_created__1 customresource_kind")
	expectEqual(t, ratelimitpolicy1CreatedLabels["customresource_version"], "v1", "gatewayapi_ratelimitpolicy_created__1 customresource_version")
	expectEqual(t, ratelimitpolicy1CreatedLabels["name"], "testratelimitpolicy1", "gatewayapi_ratelimitpolicy_created__1 name")
	expectEqual(t, ratelimitpolicy1CreatedLabels["namespace"], "default", "gatewayapi_ratelimitpolicy_created__1 namespace")

	//gatewayapi_ratelimitpolicy_target_info
	ratelimitpolicyParentInfo := metrics["gatewayapi_ratelimitpolicy_target_info"]
	ratelimitpolicy1ParentInfo1 := ratelimitpolicyParentInfo[0]
	expectEqual(t, ratelimitpolicy1ParentInfo1[3], "1", "gatewayapi_ratelimitpolicy_target_info__1 value")
	ratelimitpolicy1ParentInfo1Labels := parseLabels(string(ratelimitpolicy1ParentInfo1[2]))
	expectEqual(t, ratelimitpolicy1ParentInfo1Labels["customresource_group"], "kuadrant.io", "gatewayapi_ratelimitpolicy_target_info__1 customresource_group")
	expectEqual(t, ratelimitpolicy1ParentInfo1Labels["customresource_kind"], "RateLimitPolicy", "gatewayapi_ratelimitpolicy_target_info__1 customresource_kind")
	expectEqual(t, ratelimitpolicy1ParentInfo1Labels["customresource_version"], "v1", "gatewayapi_ratelimitpolicy_target_info__1 customresource_version")
	expectEqual(t, ratelimitpolicy1ParentInfo1Labels["name"], "testratelimitpolicy1", "gatewayapi_ratelimitpolicy_target_info__1 name")
	expectEqual(t, ratelimitpolicy1ParentInfo1Labels["namespace"], "default", "gatewayapi_ratelimitpolicy_target_info__1 namespace")
	expectEqual(t, ratelimitpolicy1ParentInfo1Labels["target_group"], "gateway.networking.k8s.io", "gatewayapi_ratelimitpolicy_target_info__1 target_group")
	expectEqual(t, ratelimitpolicy1ParentInfo1Labels["target_kind"], "HTTPRoute", "gatewayapi_ratelimitpolicy_target_info__1 target_kind")
	expectEqual(t, ratelimitpolicy1ParentInfo1Labels["target_name"], "testname1", "gatewayapi_ratelimitpolicy_target_info__1 target_name")

	//gatewayapi_ratelimitpolicy_status
	ratelimitpolicyStatus := metrics["gatewayapi_ratelimitpolicy_status"]
	ratelimitpolicy1Status1 := ratelimitpolicyStatus[0]
	expectEqual(t, ratelimitpolicy1Status1[3], "1", "gatewayapi_ratelimitpolicy_status__1 value")
	ratelimitpolicy1Status1Labels := parseLabels(string(ratelimitpolicy1Status1[2]))
	expectEqual(t, ratelimitpolicy1Status1Labels["customresource_group"], "kuadrant.io", "gatewayapi_ratelimitpolicy_status__1 customresource_group")
	expectEqual(t, ratelimitpolicy1Status1Labels["customresource_kind"], "RateLimitPolicy", "gatewayapi_ratelimitpolicy_status__1 customresource_kind")
	expectEqual(t, ratelimitpolicy1Status1Labels["customresource_version"], "v1", "gatewayapi_ratelimitpolicy_status__1 customresource_version")
	expectEqual(t, ratelimitpolicy1Status1Labels["name"], "testratelimitpolicy1", "gatewayapi_ratelimitpolicy_status__1 name")
	expectEqual(t, ratelimitpolicy1Status1Labels["namespace"], "default", "gatewayapi_ratelimitpolicy_status__1 namespace")
	expectEqual(t, ratelimitpolicy1Status1Labels["type"], "Available", "gatewayapi_ratelimitpolicy_status__1 type")
}

func testTLSPolicy(t *testing.T, metrics map[string][][]string) {
	// gatewayapi_tlspolicy_created
	tlspolicyCreated := metrics["gatewayapi_tlspolicy_created"]
	tlspolicy1Created := tlspolicyCreated[0]
	expectValidTimestampInPast(t, tlspolicy1Created[3], "gatewayapi_tlspolicy_created__1 value")
	tlspolicy1CreatedLabels := parseLabels(string(tlspolicy1Created[2]))
	expectEqual(t, tlspolicy1CreatedLabels["customresource_group"], "kuadrant.io", "gatewayapi_tlspolicy_created__1 customresource_group")
	expectEqual(t, tlspolicy1CreatedLabels["customresource_kind"], "TLSPolicy", "gatewayapi_tlspolicy_created__1 customresource_kind")
	expectEqual(t, tlspolicy1CreatedLabels["customresource_version"], "v1", "gatewayapi_tlspolicy_created__1 customresource_version")
	expectEqual(t, tlspolicy1CreatedLabels["name"], "testtlspolicy1", "gatewayapi_tlspolicy_created__1 name")
	expectEqual(t, tlspolicy1CreatedLabels["namespace"], "default", "gatewayapi_tlspolicy_created__1 namespace")

	//gatewayapi_tlspolicy_target_info
	tlspolicyParentInfo := metrics["gatewayapi_tlspolicy_target_info"]
	tlspolicy1ParentInfo1 := tlspolicyParentInfo[0]
	expectEqual(t, tlspolicy1ParentInfo1[3], "1", "gatewayapi_tlspolicy_target_info__1 value")
	tlspolicy1ParentInfo1Labels := parseLabels(string(tlspolicy1ParentInfo1[2]))
	expectEqual(t, tlspolicy1ParentInfo1Labels["customresource_group"], "kuadrant.io", "gatewayapi_tlspolicy_target_info__1 customresource_group")
	expectEqual(t, tlspolicy1ParentInfo1Labels["customresource_kind"], "TLSPolicy", "gatewayapi_tlspolicy_target_info__1 customresource_kind")
	expectEqual(t, tlspolicy1ParentInfo1Labels["customresource_version"], "v1", "gatewayapi_tlspolicy_target_info__1 customresource_version")
	expectEqual(t, tlspolicy1ParentInfo1Labels["name"], "testtlspolicy1", "gatewayapi_tlspolicy_target_info__1 name")
	expectEqual(t, tlspolicy1ParentInfo1Labels["namespace"], "default", "gatewayapi_tlspolicy_target_info__1 namespace")
	expectEqual(t, tlspolicy1ParentInfo1Labels["target_group"], "gateway.networking.k8s.io", "gatewayapi_tlspolicy_target_info__1 target_group")
	expectEqual(t, tlspolicy1ParentInfo1Labels["target_kind"], "Gateway", "gatewayapi_tlspolicy_target_info__1 target_kind")
	expectEqual(t, tlspolicy1ParentInfo1Labels["target_name"], "testgateway1", "gatewayapi_tlspolicy_target_info__1 target_name")

	//gatewayapi_tlspolicy_status
	tlspolicyStatus := metrics["gatewayapi_tlspolicy_status"]
	tlspolicy1Status1 := tlspolicyStatus[0]
	expectEqual(t, tlspolicy1Status1[3], "1", "gatewayapi_tlspolicy_status__1 value")
	tlspolicy1Status1Labels := parseLabels(string(tlspolicy1Status1[2]))
	expectEqual(t, tlspolicy1Status1Labels["customresource_group"], "kuadrant.io", "gatewayapi_tlspolicy_status__1 customresource_group")
	expectEqual(t, tlspolicy1Status1Labels["customresource_kind"], "TLSPolicy", "gatewayapi_tlspolicy_status__1 customresource_kind")
	expectEqual(t, tlspolicy1Status1Labels["customresource_version"], "v1", "gatewayapi_tlspolicy_status__1 customresource_version")
	expectEqual(t, tlspolicy1Status1Labels["name"], "testtlspolicy1", "gatewayapi_tlspolicy_status__1 name")
	expectEqual(t, tlspolicy1Status1Labels["namespace"], "default", "gatewayapi_tlspolicy_status__1 namespace")
	expectEqual(t, tlspolicy1Status1Labels["type"], "Ready", "gatewayapi_tlspolicy_status__1 type")
}

func testDNSPolicy(t *testing.T, metrics map[string][][]string) {
	// gatewayapi_dnspolicy_created
	dnspolicyCreated := metrics["gatewayapi_dnspolicy_created"]
	dnspolicy1Created := dnspolicyCreated[0]
	expectValidTimestampInPast(t, dnspolicy1Created[3], "gatewayapi_dnspolicy_created__1 value")
	dnspolicy1CreatedLabels := parseLabels(string(dnspolicy1Created[2]))
	expectEqual(t, dnspolicy1CreatedLabels["customresource_group"], "kuadrant.io", "gatewayapi_dnspolicy_created__1 customresource_group")
	expectEqual(t, dnspolicy1CreatedLabels["customresource_kind"], "DNSPolicy", "gatewayapi_dnspolicy_created__1 customresource_kind")
	expectEqual(t, dnspolicy1CreatedLabels["customresource_version"], "v1", "gatewayapi_dnspolicy_created__1 customresource_version")
	expectEqual(t, dnspolicy1CreatedLabels["name"], "testdnspolicy1", "gatewayapi_dnspolicy_created__1 name")
	expectEqual(t, dnspolicy1CreatedLabels["namespace"], "default", "gatewayapi_dnspolicy_created__1 namespace")

	//gatewayapi_dnspolicy_target_info
	dnspolicyParentInfo := metrics["gatewayapi_dnspolicy_target_info"]
	dnspolicy1ParentInfo1 := dnspolicyParentInfo[0]
	expectEqual(t, dnspolicy1ParentInfo1[3], "1", "gatewayapi_dnspolicy_target_info__1 value")
	dnspolicy1ParentInfo1Labels := parseLabels(string(dnspolicy1ParentInfo1[2]))
	expectEqual(t, dnspolicy1ParentInfo1Labels["customresource_group"], "kuadrant.io", "gatewayapi_dnspolicy_target_info__1 customresource_group")
	expectEqual(t, dnspolicy1ParentInfo1Labels["customresource_kind"], "DNSPolicy", "gatewayapi_dnspolicy_target_info__1 customresource_kind")
	expectEqual(t, dnspolicy1ParentInfo1Labels["customresource_version"], "v1", "gatewayapi_dnspolicy_target_info__1 customresource_version")
	expectEqual(t, dnspolicy1ParentInfo1Labels["name"], "testdnspolicy1", "gatewayapi_dnspolicy_target_info__1 name")
	expectEqual(t, dnspolicy1ParentInfo1Labels["namespace"], "default", "gatewayapi_dnspolicy_target_info__1 namespace")
	expectEqual(t, dnspolicy1ParentInfo1Labels["target_group"], "gateway.networking.k8s.io", "gatewayapi_dnspolicy_target_info__1 target_group")
	expectEqual(t, dnspolicy1ParentInfo1Labels["target_kind"], "Gateway", "gatewayapi_dnspolicy_target_info__1 target_kind")
	expectEqual(t, dnspolicy1ParentInfo1Labels["target_name"], "testgateway1", "gatewayapi_dnspolicy_target_info__1 target_name")

	//gatewayapi_dnspolicy_status
	dnspolicyStatus := metrics["gatewayapi_dnspolicy_status"]
	dnspolicy1Status1 := dnspolicyStatus[0]
	expectEqual(t, dnspolicy1Status1[3], "1", "gatewayapi_dnspolicy_status__1 value")
	dnspolicy1Status1Labels := parseLabels(string(dnspolicy1Status1[2]))
	expectEqual(t, dnspolicy1Status1Labels["customresource_group"], "kuadrant.io", "gatewayapi_dnspolicy_status__1 customresource_group")
	expectEqual(t, dnspolicy1Status1Labels["customresource_kind"], "DNSPolicy", "gatewayapi_dnspolicy_status__1 customresource_kind")
	expectEqual(t, dnspolicy1Status1Labels["customresource_version"], "v1", "gatewayapi_dnspolicy_status__1 customresource_version")
	expectEqual(t, dnspolicy1Status1Labels["name"], "testdnspolicy1", "gatewayapi_dnspolicy_status__1 name")
	expectEqual(t, dnspolicy1Status1Labels["namespace"], "default", "gatewayapi_dnspolicy_status__1 namespace")
	expectEqual(t, dnspolicy1Status1Labels["type"], "Ready", "gatewayapi_dnspolicy_status__1 type")
}

func testAuthPolicy(t *testing.T, metrics map[string][][]string) {
	// gatewayapi_authpolicy_created
	authpolicyCreated := metrics["gatewayapi_authpolicy_created"]
	authpolicy1Created := authpolicyCreated[0]
	expectValidTimestampInPast(t, authpolicy1Created[3], "gatewayapi_authpolicy_created__1 value")
	authpolicy1CreatedLabels := parseLabels(string(authpolicy1Created[2]))
	expectEqual(t, authpolicy1CreatedLabels["customresource_group"], "kuadrant.io", "gatewayapi_authpolicy_created__1 customresource_group")
	expectEqual(t, authpolicy1CreatedLabels["customresource_kind"], "AuthPolicy", "gatewayapi_authpolicy_created__1 customresource_kind")
	expectEqual(t, authpolicy1CreatedLabels["customresource_version"], "v1", "gatewayapi_authpolicy_created__1 customresource_version")
	expectEqual(t, authpolicy1CreatedLabels["name"], "testauthpolicy1", "gatewayapi_authpolicy_created__1 name")
	expectEqual(t, authpolicy1CreatedLabels["namespace"], "default", "gatewayapi_authpolicy_created__1 namespace")

	//gatewayapi_authpolicy_target_info
	authpolicyParentInfo := metrics["gatewayapi_authpolicy_target_info"]
	authpolicy1ParentInfo1 := authpolicyParentInfo[0]
	expectEqual(t, authpolicy1ParentInfo1[3], "1", "gatewayapi_authpolicy_target_info__1 value")
	authpolicy1ParentInfo1Labels := parseLabels(string(authpolicy1ParentInfo1[2]))
	expectEqual(t, authpolicy1ParentInfo1Labels["customresource_group"], "kuadrant.io", "gatewayapi_authpolicy_target_info__1 customresource_group")
	expectEqual(t, authpolicy1ParentInfo1Labels["customresource_kind"], "AuthPolicy", "gatewayapi_authpolicy_target_info__1 customresource_kind")
	expectEqual(t, authpolicy1ParentInfo1Labels["customresource_version"], "v1", "gatewayapi_authpolicy_target_info__1 customresource_version")
	expectEqual(t, authpolicy1ParentInfo1Labels["name"], "testauthpolicy1", "gatewayapi_authpolicy_target_info__1 name")
	expectEqual(t, authpolicy1ParentInfo1Labels["namespace"], "default", "gatewayapi_authpolicy_target_info__1 namespace")
	expectEqual(t, authpolicy1ParentInfo1Labels["target_group"], "gateway.networking.k8s.io", "gatewayapi_authpolicy_target_info__1 target_group")
	expectEqual(t, authpolicy1ParentInfo1Labels["target_kind"], "HTTPRoute", "gatewayapi_authpolicy_target_info__1 target_kind")
	expectEqual(t, authpolicy1ParentInfo1Labels["target_name"], "testgateway1", "gatewayapi_authpolicy_target_info__1 target_name")

	//gatewayapi_authpolicy_status
	authpolicyStatus := metrics["gatewayapi_authpolicy_status"]
	authpolicy1Status1 := authpolicyStatus[0]
	expectEqual(t, authpolicy1Status1[3], "1", "gatewayapi_authpolicy_status__1 value")
	authpolicy1Status1Labels := parseLabels(string(authpolicy1Status1[2]))
	expectEqual(t, authpolicy1Status1Labels["customresource_group"], "kuadrant.io", "gatewayapi_authpolicy_status__1 customresource_group")
	expectEqual(t, authpolicy1Status1Labels["customresource_kind"], "AuthPolicy", "gatewayapi_authpolicy_status__1 customresource_kind")
	expectEqual(t, authpolicy1Status1Labels["customresource_version"], "v1", "gatewayapi_authpolicy_status__1 customresource_version")
	expectEqual(t, authpolicy1Status1Labels["name"], "testauthpolicy1", "gatewayapi_authpolicy_status__1 name")
	expectEqual(t, authpolicy1Status1Labels["namespace"], "default", "gatewayapi_authpolicy_status__1 namespace")
	expectEqual(t, authpolicy1Status1Labels["type"], "Available", "gatewayapi_authpolicy_status__1 type")
}

func testDNSRecord(t *testing.T, metrics map[string][][]string) {
	// kuadrant_dnsrecord_created
	dnsrecordCreated := metrics["kuadrant_dnsrecord_created"]
	dnsrecord1Created := dnsrecordCreated[0]
	expectValidTimestampInPast(t, dnsrecord1Created[3], "kuadrant_dnsrecord_created__1 value")
	dnsrecord1CreatedLabels := parseLabels(string(dnsrecord1Created[2]))
	expectEqual(t, dnsrecord1CreatedLabels["customresource_group"], "kuadrant.io", "kuadrant_dnsrecord_created__1 customresource_group")
	expectEqual(t, dnsrecord1CreatedLabels["customresource_kind"], "DNSRecord", "kuadrant_dnsrecord_created__1 customresource_kind")
	expectEqual(t, dnsrecord1CreatedLabels["customresource_version"], "v1alpha1", "kuadrant_dnsrecord_created__1 customresource_version")
	expectEqual(t, dnsrecord1CreatedLabels["name"], "testdnsrecord1", "kuadrant_dnsrecord_created__1 name")
	expectEqual(t, dnsrecord1CreatedLabels["namespace"], "default", "kuadrant_dnsrecord_created__1 namespace")

	//kuadrant_dnsrecord_status
	dnsrecordStatus := metrics["kuadrant_dnsrecord_status"]
	dnsrecord1Status1 := dnsrecordStatus[0]
	expectEqual(t, dnsrecord1Status1[3], "1", "kuadrant_dnsrecord_status__1 value")
	dnsrecord1Status1Labels := parseLabels(string(dnsrecord1Status1[2]))
	expectEqual(t, dnsrecord1Status1Labels["customresource_group"], "kuadrant.io", "kuadrant_dnsrecord_status__1 customresource_group")
	expectEqual(t, dnsrecord1Status1Labels["customresource_kind"], "DNSRecord", "kuadrant_dnsrecord_status__1 customresource_kind")
	expectEqual(t, dnsrecord1Status1Labels["customresource_version"], "v1alpha1", "kuadrant_dnsrecord_status__1 customresource_version")
	expectEqual(t, dnsrecord1Status1Labels["name"], "testdnsrecord1", "kuadrant_dnsrecord_status__1 name")
	expectEqual(t, dnsrecord1Status1Labels["namespace"], "default", "kuadrant_dnsrecord_status__1 namespace")
	expectEqual(t, dnsrecord1Status1Labels["type"], "Ready", "kuadrant_dnsrecord_status__1 type")

	//kuadrant_dnsrecord_status_root_domain_owners
	dnsrecordStatusRootDomainOwners := metrics["kuadrant_dnsrecord_status_root_domain_owners"]
	dnsrecord1StatusRootDomainOwners1 := dnsrecordStatusRootDomainOwners[0]
	expectEqual(t, dnsrecord1StatusRootDomainOwners1[3], "1", "kuadrant_dnsrecord_status_root_domain_owners__1 value")
	dnsrecord1StatusRootDomainOwners1Labels := parseLabels(string(dnsrecord1StatusRootDomainOwners1[2]))
	expectEqual(t, dnsrecord1StatusRootDomainOwners1Labels["customresource_group"], "kuadrant.io", "kuadrant_dnsrecord_status_root_domain_owners__1 customresource_group")
	expectEqual(t, dnsrecord1StatusRootDomainOwners1Labels["customresource_kind"], "DNSRecord", "kuadrant_dnsrecord_status_root_domain_owners__1 customresource_kind")
	expectEqual(t, dnsrecord1StatusRootDomainOwners1Labels["customresource_version"], "v1alpha1", "kuadrant_dnsrecord_status_root_domain_owners__1 customresource_version")
	expectEqual(t, dnsrecord1StatusRootDomainOwners1Labels["name"], "testdnsrecord1", "kuadrant_dnsrecord_status_root_domain_owners__1 name")
	expectEqual(t, dnsrecord1StatusRootDomainOwners1Labels["namespace"], "default", "kuadrant_dnsrecord_status_root_domain_owners__1 namespace")

	expectedRootDomainOwners := map[int]string{
		0: "k4ww8e00",
		1: "mvg80cg8",
	}

	for i, rootDomainOwner := range dnsrecordStatusRootDomainOwners {
		rootDomainOwnerInfo := parseLabels(string(rootDomainOwner[0]))
		rootDomainOwnerName := rootDomainOwnerInfo["owner"]
		expectEqual(t, rootDomainOwnerName, expectedRootDomainOwners[i], "kuadrant_dnsrecord_status_root_domain_owners__"+strconv.Itoa(i)+" owner")
	}
}

func parseLabels(labelsRaw string) map[string]string {
	// simple label parsing assuming no special chars/escaping
	// fmt.Printf("labelsRaw=%s\n", labelsRaw)
	labels := map[string]string{}
	labelParts := strings.Split(labelsRaw, ",")
	// fmt.Printf("labelParts=%v\n", labelParts)
	for _, labelPart := range labelParts {
		labelNameVal := strings.Split(labelPart, "=")
		// fmt.Printf("labelNameVal=%v\n", labelNameVal)
		labels[labelNameVal[0]] = labelNameVal[1][1 : len(labelNameVal[1])-1]
	}
	return labels
}

func expectEqual(t *testing.T, actual string, expected string, msg string) {
	if actual != expected {
		t.Fatalf("(%s) Expected %s to equal %s", msg, actual, expected)
	}
}

func expectValidTimestampInPast(t *testing.T, timestamp string, msg string) {
	flt, err := strconv.ParseFloat(timestamp, 64)
	if err != nil {
		t.Fatalf("(%s) Failed parsing timestamp %s", msg, timestamp)
	}
	if flt < 1 || flt > float64(time.Now().Unix()) {
		t.Fatalf("(%s) Expected a valid timestamp in the past, but got value of %s", msg, timestamp)
	}
}
