import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:sandwitches/src/telemetry/telemetry.dart';

/// Records requests and replies with a canned body — no real network involved.
class _FakeAdapter implements HttpClientAdapter {
  _FakeAdapter(this.body, {this.statusCode = 200});

  final String body;
  final int statusCode;
  final List<RequestOptions> requests = [];

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);
    return ResponseBody.fromString(
      body,
      statusCode,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}

void main() {
  group('normalizeEndpoint', () {
    test('appends the OTLP traces path', () {
      expect(
        Telemetry.normalizeEndpoint('http://collector:4318'),
        'http://collector:4318/v1/traces',
      );
    });

    test('strips trailing slashes', () {
      expect(
        Telemetry.normalizeEndpoint('http://collector:4318///'),
        'http://collector:4318/v1/traces',
      );
    });

    test('preserves an endpoint that already has the path', () {
      expect(
        Telemetry.normalizeEndpoint('https://otlp.example.com/v1/traces'),
        'https://otlp.example.com/v1/traces',
      );
    });
  });

  group('parseOtelHeaders', () {
    test('parses comma separated key=value pairs', () {
      expect(
        Telemetry.parseOtelHeaders('x-api-key=secret, x-tenant=acme'),
        {'x-api-key': 'secret', 'x-tenant': 'acme'},
      );
    });

    test('URL-decodes values and skips malformed pairs', () {
      expect(
        Telemetry.parseOtelHeaders('auth=Bearer%20abc,bogus,=novalue,empty='),
        {'auth': 'Bearer abc'},
      );
    });

    test('returns an empty map for empty input', () {
      expect(Telemetry.parseOtelHeaders(''), isEmpty);
    });
  });

  group('buildOtlpPayload', () {
    test('has resource, scope and span shape', () {
      final span = TraceSpan(
        name: 'app.start',
        attributes: const {'screen.name': 'home'},
        traceId: '4bf92f3577b34da6a3ce929d0e0e4736',
        spanId: '00f067aa0ba902b7',
        startNanos: 1,
      );
      final payload = buildOtlpPayload(
        serviceName: 'sandwitches-android',
        serviceVersion: '1.0.0',
        span: span,
        attributes: span.attributes,
        failed: false,
        endNanos: 2,
      );

      final resourceSpans = payload['resourceSpans'] as List;
      final resource = resourceSpans.single as Map;
      final resourceAttrs = (resource['resource'] as Map)['attributes'] as List;
      String? attrValue(String key) {
        for (final entry in resourceAttrs) {
          final map = entry as Map;
          if (map['key'] == key) {
            return ((map['value'] as Map)['stringValue'] as String?) ;
          }
        }
        return null;
      }

      expect(attrValue('service.name'), 'sandwitches-android');
      expect(attrValue('telemetry.sdk.name'), 'sandwitches-mobile');
      expect(attrValue('telemetry.sdk.language'), 'dart');

      final scope = (resource['scopeSpans'] as List).single as Map;
      expect(
        scope['scope'],
        {'name': 'sandwitches.mobile', 'version': '1.0.0'},
      );
      final exported = (scope['spans'] as List).single as Map;
      expect(exported['traceId'], span.traceId);
      expect(exported['spanId'], span.spanId);
      expect(exported['name'], 'app.start');
      expect(exported['startTimeUnixNano'], '1');
      expect(exported['endTimeUnixNano'], '2');
      expect(exported['status'], {'code': 1});
      final spanAttrs = exported['attributes'] as List;
      expect(
        spanAttrs.map((e) => (e as Map)['key']),
        contains('screen.name'),
      );
    });

    test('marks a failed span with error status', () {
      final span = TraceSpan(
        name: 'http.request',
        attributes: const {},
        traceId: '4bf92f3577b34da6a3ce929d0e0e4736',
        spanId: '00f067aa0ba902b7',
        startNanos: 1,
      );
      final payload = buildOtlpPayload(
        serviceName: 'sandwitches-android',
        serviceVersion: '1.0.0',
        span: span,
        attributes: const {},
        failed: true,
        endNanos: 2,
      );
      final resource =
          (payload['resourceSpans'] as List).single as Map;
      final scope = (resource['scopeSpans'] as List).single as Map;
      final exported = (scope['spans'] as List).single as Map;
      expect(exported['status'], {'code': 2});
    });
  });

  group('OtelSettings.fromJson', () {
    test('accepts otel_ prefixed keys', () {
      final settings = OtelSettings.fromJson({
        'otel_endpoint': 'http://collector:4318',
        'otel_enabled': true,
      });
      expect(settings.endpoint, 'http://collector:4318');
      expect(settings.enabled, isTrue);
      expect(settings.tracesEnabled, isTrue);
      expect(settings.metricsEnabled, isTrue);
      expect(settings.logsEnabled, isTrue);
    });

    test('accepts plain keys', () {
      final settings = OtelSettings.fromJson({
        'endpoint': 'http://collector:4318',
        'enabled': true,
        'traces_enabled': false,
      });
      expect(settings.endpoint, 'http://collector:4318');
      expect(settings.enabled, isTrue);
      expect(settings.tracesEnabled, isFalse);
    });

    test('traces defaults to enabled but explicit false wins', () {
      final defaults = OtelSettings.fromJson({
        'otel_endpoint': 'http://collector:4318',
        'otel_enabled': true,
      });
      expect(defaults.tracesEnabled, isTrue);

      final explicit = OtelSettings.fromJson({
        'otel_endpoint': 'http://collector:4318',
        'otel_enabled': true,
        'traces_enabled': false,
      });
      expect(explicit.enabled, isTrue);
      expect(explicit.tracesEnabled, isFalse);
    });

    test('an empty endpoint means disabled', () {
      final settings = OtelSettings.fromJson({
        'otel_endpoint': '',
        'otel_enabled': true,
      });
      expect(settings.endpoint, isEmpty);
      expect(settings.enabled, isFalse);
      expect(settings.tracesEnabled, isFalse);
    });

    test('a missing endpoint means disabled', () {
      final settings = OtelSettings.fromJson({'otel_enabled': true});
      expect(settings.enabled, isFalse);
      expect(settings.tracesEnabled, isFalse);
    });
  });

  group('fetchOtelSettings', () {
    test('parses the public settings endpoint response', () async {
      final adapter = _FakeAdapter(
        '{"otel_endpoint":"http://collector:4318","otel_enabled":true}',
      );
      final dio = Dio()..httpClientAdapter = adapter;

      final settings = await fetchOtelSettings(
        'https://sandwitches.example/',
        dio: dio,
      );

      expect(settings, isNotNull);
      expect(settings!.endpoint, 'http://collector:4318');
      expect(settings.enabled, isTrue);
      expect(settings.tracesEnabled, isTrue);
      expect(adapter.requests.single.path, contains('/api/v1/settings'));
      expect(adapter.requests.single.path, isNot(contains('//api')));
    });

    test('returns null on a non-200 response', () async {
      final dio = Dio()..httpClientAdapter = _FakeAdapter('{}', statusCode: 500);
      expect(
        await fetchOtelSettings('https://sandwitches.example', dio: dio),
        isNull,
      );
    });

    test('returns null on a transport error instead of throwing', () async {
      final dio = Dio()..httpClientAdapter = _FakeAdapter('not json');
      expect(
        await fetchOtelSettings('https://sandwitches.example', dio: dio),
        isNull,
      );
    });

    test('returns null for an empty base URL without a request', () async {
      final adapter = _FakeAdapter('{}');
      final dio = Dio()..httpClientAdapter = adapter;
      expect(await fetchOtelSettings('   ', dio: dio), isNull);
      expect(adapter.requests, isEmpty);
    });
  });
}
