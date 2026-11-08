import 'dart:async';
import 'dart:convert';
import 'dart:math';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';

/// Configuration is compile-time so secrets never live in app storage.
///
/// The OTLP endpoint can also be supplied at runtime by the server (see
/// [fetchOtelSettings]); the dart-define values are the fallback.
abstract final class TelemetryConfig {
  static const otelEndpoint = String.fromEnvironment(
    'OTEL_EXPORTER_OTLP_ENDPOINT',
  );
  static const otelHeaders = String.fromEnvironment(
    'OTEL_EXPORTER_OTLP_HEADERS',
  );
  static const serviceName = String.fromEnvironment(
    'OTEL_SERVICE_NAME',
    defaultValue: 'sandwitches-android',
  );
  static const serviceVersion = String.fromEnvironment(
    'OTEL_SERVICE_VERSION',
    defaultValue: '1.0.0',
  );
  static const umamiUrl = String.fromEnvironment('UMAMI_URL');
  static const umamiWebsiteId = String.fromEnvironment('UMAMI_WEBSITE_ID');

  static bool get otelEnabled => otelEndpoint.trim().isNotEmpty && !kIsWeb;
  static bool get umamiEnabled =>
      umamiUrl.trim().isNotEmpty && umamiWebsiteId.trim().isNotEmpty;
}

/// Runtime OpenTelemetry configuration as served by the API.
///
/// Parsing is deliberately tolerant: the server may spell the keys either
/// plainly (`endpoint`) or with the `otel_` prefix (`otel_endpoint`), and both
/// spellings are accepted here.
class OtelSettings {
  const OtelSettings({
    required this.endpoint,
    required this.enabled,
    required this.tracesEnabled,
    required this.metricsEnabled,
    required this.logsEnabled,
  });

  const OtelSettings.disabled()
    : endpoint = '',
      enabled = false,
      tracesEnabled = false,
      metricsEnabled = false,
      logsEnabled = false;

  final String endpoint;
  final bool enabled;
  final bool tracesEnabled;
  final bool metricsEnabled;
  final bool logsEnabled;

  factory OtelSettings.fromJson(Map<String, dynamic> json) {
    String pickString(List<String> keys) {
      for (final key in keys) {
        final value = json[key];
        if (value is String) return value.trim();
      }
      return '';
    }

    bool? pickBool(List<String> keys) {
      for (final key in keys) {
        final value = json[key];
        if (value is bool) return value;
        if (value is num) return value != 0;
        if (value is String) {
          final normalized = value.trim().toLowerCase();
          if (normalized == 'true' || normalized == '1') return true;
          if (normalized == 'false' || normalized == '0') return false;
        }
      }
      return null;
    }

    final endpoint = pickString(['endpoint', 'otel_endpoint']);
    // An empty endpoint means there is nowhere to export to, so the whole
    // configuration is treated as disabled regardless of the flag.
    final enabled = (pickBool(['enabled', 'otel_enabled']) ?? false) &&
        endpoint.isNotEmpty;
    return OtelSettings(
      endpoint: endpoint,
      enabled: enabled,
      tracesEnabled: (pickBool(['traces_enabled']) ?? enabled) && enabled,
      metricsEnabled: (pickBool(['metrics_enabled']) ?? enabled) && enabled,
      logsEnabled: (pickBool(['logs_enabled']) ?? enabled) && enabled,
    );
  }
}

/// Fetches runtime OpenTelemetry settings from the app's own API.
///
/// This repo already exposes its settings publicly at `GET /api/v1/settings`
/// (no auth), so that existing route is used instead of a standardized
/// `/api/otel-settings` endpoint. Never throws: any error, timeout or non-200
/// response yields `null` and telemetry simply stays on its dart-define
/// fallback (or disabled).
Future<OtelSettings?> fetchOtelSettings(
  String baseUrl, {
  Dio? dio,
  Duration timeout = const Duration(seconds: 3),
}) async {
  final base = baseUrl.trim().replaceAll(RegExp(r'/+$'), '');
  if (base.isEmpty) return null;
  final client = dio ?? Dio();
  try {
    final response = await client
        .get<dynamic>(
          '$base/api/v1/settings',
          options: Options(
            responseType: ResponseType.json,
            sendTimeout: timeout,
            receiveTimeout: timeout,
          ),
        )
        .timeout(timeout);
    if (response.statusCode != 200) return null;
    final data = response.data;
    if (data is Map) {
      return OtelSettings.fromJson(Map<String, dynamic>.from(data));
    }
    if (data is String && data.trim().isNotEmpty) {
      final decoded = jsonDecode(data);
      if (decoded is Map) {
        return OtelSettings.fromJson(Map<String, dynamic>.from(decoded));
      }
    }
    return null;
  } catch (_) {
    return null;
  }
}

/// Allowlisted analytics events. Keys and values are enforced at the boundary,
/// so recipe titles, comments, searches and user input can never be sent.
const Map<String, Set<String>> umamiEventAllowlist = {
  'screen_view': {'screen'},
  'recipe_search': {'has_query'},
  'recipe_submitted': {'mode'},
  'order_submitted': {},
  'sync_started': {},
  'theme_changed': {'theme'},
};

const Set<String> _screens = {
  'startup',
  'onboarding',
  'login',
  'home',
  'recipe',
  'recipe_create',
  'cooking',
  'cart',
  'orders',
  'order_tracker',
  'admin',
  'settings',
};
const Set<String> _themes = {'system', 'light', 'dark', 'eink'};
const Set<String> _modes = {'create', 'edit'};

Map<String, Object>? sanitizeUmamiEvent(
  String event, [
  Map<String, Object?> props = const {},
]) {
  final allowed = umamiEventAllowlist[event];
  if (allowed == null) return null;
  final clean = <String, Object>{};
  for (final entry in props.entries) {
    if (!allowed.contains(entry.key)) continue;
    final value = entry.value;
    if (value is bool || value is num) {
      clean[entry.key] = value as Object;
      continue;
    }
    if (value is! String || value.length > 64) continue;
    if (entry.key == 'screen' && !_screens.contains(value)) continue;
    if (entry.key == 'theme' && !_themes.contains(value)) continue;
    if (entry.key == 'mode' && !_modes.contains(value)) continue;
    clean[entry.key] = value;
  }
  if (allowed.isNotEmpty && clean.isEmpty) return null;
  return clean;
}

String randomHex(int bytes) {
  final random = Random.secure();
  final buffer = StringBuffer();
  for (var i = 0; i < bytes; i++) {
    buffer.write(random.nextInt(256).toRadixString(16).padLeft(2, '0'));
  }
  return buffer.toString();
}

String buildTraceparent(String traceId, String spanId) =>
    '00-$traceId-$spanId-01';

class TraceSpan {
  TraceSpan({
    required this.name,
    required this.attributes,
    required this.traceId,
    required this.spanId,
    required this.startNanos,
    this.kind = 1,
  });

  final String name;
  final Map<String, Object> attributes;
  final String traceId;
  final String spanId;
  final int startNanos;
  final int kind;
  bool _ended = false;
  int? statusCode;
  String error = '';

  String get traceparent => buildTraceparent(traceId, spanId);

  void end({
    int? httpStatus,
    Object? failure,
    Map<String, Object> extra = const {},
  }) {
    if (_ended) return;
    _ended = true;
    statusCode = httpStatus;
    if (failure != null) error = failure.toString();
    final attrs = <String, Object>{...attributes, ...extra};
    if (httpStatus != null) attrs['http.response.status_code'] = httpStatus;
    if (error.isNotEmpty) attrs['error.message'] = _short(error);
    Telemetry.instance._finish(
      this,
      attrs,
      error.isNotEmpty || (httpStatus != null && httpStatus >= 500),
    );
  }

  String _short(String value) =>
      value.length > 200 ? '${value.substring(0, 200)}…' : value;
}

Map<String, Object?> buildOtlpPayload({
  required String serviceName,
  required String serviceVersion,
  required TraceSpan span,
  required Map<String, Object> attributes,
  required bool failed,
  required int endNanos,
}) {
  List<Map<String, Object?>> attrs(Map<String, Object> values) => values.entries
      .map(
        (entry) => {
          'key': entry.key,
          'value': switch (entry.value) {
            final bool value => {'boolValue': value},
            final int value => {'intValue': value.toString()},
            final double value => {'doubleValue': value},
            _ => {'stringValue': entry.value.toString()},
          },
        },
      )
      .toList();

  return {
    'resourceSpans': [
      {
        'resource': {
          'attributes': attrs({
            'service.name': serviceName,
            'service.version': serviceVersion,
            'telemetry.sdk.name': 'sandwitches-mobile',
            'telemetry.sdk.language': 'dart',
          }),
        },
        'scopeSpans': [
          {
            'scope': {'name': 'sandwitches.mobile', 'version': serviceVersion},
            'spans': [
              {
                'traceId': span.traceId,
                'spanId': span.spanId,
                'name': span.name,
                'kind': span.kind,
                'startTimeUnixNano': span.startNanos.toString(),
                'endTimeUnixNano': endNanos.toString(),
                'attributes': attrs(attributes),
                'status': {'code': failed ? 2 : 1},
              },
            ],
          },
        ],
      },
    ],
  };
}

class Telemetry {
  Telemetry._();

  static final Telemetry instance = Telemetry._();

  Dio? _dio;
  Timer? _timer;
  final List<Map<String, Object?>> _queue = [];
  bool _started = false;
  String _tracesEndpoint = '';
  Map<String, String> _headers = const {};

  bool get otelEnabled => _tracesEndpoint.isNotEmpty;
  bool get umamiEnabled => TelemetryConfig.umamiEnabled;

  /// Starts the exporter.
  ///
  /// Uses [endpoint] when supplied, otherwise the `OTEL_EXPORTER_OTLP_ENDPOINT`
  /// dart-define. With neither, telemetry stays disabled. Calling this more
  /// than once is a no-op.
  Future<void> start({String? endpoint, Map<String, String>? headers}) async {
    if (_started) return;
    _started = true;
    if (kIsWeb) return;
    final raw = (endpoint != null && endpoint.trim().isNotEmpty)
        ? endpoint
        : TelemetryConfig.otelEndpoint;
    if (raw.trim().isEmpty) return;
    _dio = Dio(
      BaseOptions(
        connectTimeout: const Duration(seconds: 10),
        receiveTimeout: const Duration(seconds: 10),
        validateStatus: (status) => status != null && status < 500,
      ),
    );
    _tracesEndpoint = normalizeEndpoint(raw);
    _headers = headers ?? parseOtelHeaders(TelemetryConfig.otelHeaders);
    _timer = Timer.periodic(const Duration(seconds: 5), (_) => _flush());
    final span = beginSpan('app.start');
    span.end();
  }

  /// Resolves the OTLP endpoint from the server's public settings endpoint and
  /// starts exporting when it is enabled; otherwise falls back to the
  /// dart-define configuration (which may itself leave telemetry disabled).
  Future<void> startFromServer(String baseUrl, {Dio? dio}) async {
    if (_started) return;
    OtelSettings? settings;
    try {
      settings = await fetchOtelSettings(baseUrl, dio: dio);
    } catch (_) {
      settings = null;
    }
    if (settings != null && settings.enabled && settings.tracesEnabled) {
      await start(endpoint: settings.endpoint);
      return;
    }
    await start();
  }

  TraceSpan beginSpan(
    String name, {
    Map<String, Object> attributes = const {},
    int kind = 1,
  }) {
    return TraceSpan(
      name: name,
      attributes: attributes,
      traceId: randomHex(16),
      spanId: randomHex(8),
      startNanos: DateTime.now().microsecondsSinceEpoch * 1000,
      kind: kind,
    );
  }

  void screen(String name) {
    final span = beginSpan('screen $name', attributes: {'screen.name': name});
    span.end();
    track('screen_view', {'screen': name});
  }

  void track(String event, [Map<String, Object?> props = const {}]) {
    final clean = sanitizeUmamiEvent(event, props);
    if (clean == null || !TelemetryConfig.umamiEnabled) return;
    final client = _dio ??= Dio();
    final base = TelemetryConfig.umamiUrl.replaceAll(RegExp(r'/+$'), '');
    final uri = Uri.tryParse('$base/api/send');
    if (uri == null) return;
    final body = {
      'type': 'event',
      'payload': {
        'website': TelemetryConfig.umamiWebsiteId,
        'hostname': 'sandwitches.app',
        'url': '/app/$event',
        'event': event,
        'data': clean,
      },
    };
    unawaited(
      client
          .post<dynamic>(
            uri.toString(),
            data: jsonEncode(body),
            options: Options(
              contentType: Headers.jsonContentType,
              validateStatus: (status) => status != null && status < 500,
            ),
          )
          .then((_) {}, onError: (Object _) {}),
    );
  }

  void _finish(TraceSpan span, Map<String, Object> attributes, bool failed) {
    if (!otelEnabled) return;
    final payload = buildOtlpPayload(
      serviceName: TelemetryConfig.serviceName,
      serviceVersion: TelemetryConfig.serviceVersion,
      span: span,
      attributes: attributes,
      failed: failed,
      endNanos: DateTime.now().microsecondsSinceEpoch * 1000,
    );
    _queue.add(payload);
    if (_queue.length >= 20) _flush();
  }

  void _flush() {
    if (_queue.isEmpty || _dio == null || _tracesEndpoint.isEmpty) return;
    final spans = <Map<String, Object?>>[];
    for (final payload in _queue) {
      final resourceSpans = payload['resourceSpans'];
      if (resourceSpans is List && resourceSpans.isNotEmpty) {
        final resource = resourceSpans.first;
        if (resource is Map) {
          final scopeSpans = resource['scopeSpans'];
          if (scopeSpans is List && scopeSpans.isNotEmpty) {
            final scope = scopeSpans.first;
            if (scope is Map && scope['spans'] is List) {
              spans.addAll(
                (scope['spans'] as List).cast<Map<String, Object?>>(),
              );
            }
          }
        }
      }
    }
    _queue.clear();
    if (spans.isEmpty) return;
    final payload = {
      'resourceSpans': [
        {
          'resource': {
            'attributes': [
              {
                'key': 'service.name',
                'value': {'stringValue': TelemetryConfig.serviceName},
              },
              {
                'key': 'service.version',
                'value': {'stringValue': TelemetryConfig.serviceVersion},
              },
            ],
          },
          'scopeSpans': [
            {
              'scope': {'name': 'sandwitches.mobile'},
              'spans': spans,
            },
          ],
        },
      ],
    };
    unawaited(
      _dio!
          .post<dynamic>(
            _tracesEndpoint,
            data: jsonEncode(payload),
            options: Options(
              contentType: Headers.jsonContentType,
              headers: {..._headers},
            ),
          )
          .then(
            (response) {
              final code = response.statusCode ?? 0;
              if (code >= 400 && kDebugMode) {
                debugPrint('OTLP export failed: $code');
              }
            },
            onError: (Object error) {
              if (kDebugMode) debugPrint('OTLP export failed: $error');
            },
          ),
    );
  }

  Future<void> shutdown() async {
    _timer?.cancel();
    _flush();
    await Future<void>.delayed(const Duration(milliseconds: 200));
    _dio?.close();
    _dio = null;
    _started = false;
    _tracesEndpoint = '';
  }

  /// Appends the OTLP/HTTP traces path unless the endpoint already ends with
  /// it. Trailing slashes are removed first.
  static String normalizeEndpoint(String raw) {
    final base = raw.trim().replaceAll(RegExp(r'/+$'), '');
    if (base.isEmpty) return '';
    if (base.endsWith('/v1/traces')) return base;
    return '$base/v1/traces';
  }

  /// Parses the `key=value,key=value` form of `OTEL_EXPORTER_OTLP_HEADERS`,
  /// URL-decoding each value.
  static Map<String, String> parseOtelHeaders(String raw) {
    final out = <String, String>{};
    for (final pair in raw.split(',')) {
      final idx = pair.indexOf('=');
      if (idx <= 0) continue;
      final key = pair.substring(0, idx).trim();
      var value = pair.substring(idx + 1).trim();
      try {
        value = Uri.decodeComponent(value);
      } catch (_) {}
      if (key.isNotEmpty && value.isNotEmpty) out[key] = value;
    }
    return out;
  }
}
