import 'package:dio/dio.dart';
import 'package:dio_cookie_manager/dio_cookie_manager.dart';
import 'package:cookie_jar/cookie_jar.dart';

import 'models.dart';
import 'providers.dart' show ApiException;

class ApiClient {
  final Dio dio;
  final PersistCookieJar cookieJar;
  final String baseUrl;
  ApiClient({required this.baseUrl, required this.cookieJar}) : dio = Dio(BaseOptions(
    baseUrl: baseUrl.endsWith('/') ? baseUrl.substring(0, baseUrl.length-1) : baseUrl,
    connectTimeout: const Duration(seconds: 10),
    receiveTimeout: const Duration(seconds: 10),
    followRedirects: false,
    validateStatus: (s) => s != null && s < 400,
  )) {
    dio.interceptors.add(CookieManager(cookieJar));
  }

  String get apiBase => '${dio.options.baseUrl}/api/v1';

  Future<bool> ping() async {
    try {
      final r = await dio.get('$apiBase/ping');
      return r.statusCode != null && r.statusCode! < 400;
    } catch (_) { return false; }
  }

  Future<Response> login(String username, String password) async {
    return dio.post('$baseUrl/login',
      data: {'username': username, 'password': password},
      options: Options(contentType: Headers.formUrlEncodedContentType, followRedirects: false, validateStatus: (s)=> s!=null && s<500),
    );
  }

  Future<Map<String,dynamic>?> fetchMe() async {
    try {
      final r = await dio.get('$apiBase/me');
      if (r.statusCode == 200 && r.data is Map) return r.data as Map<String,dynamic>;
      return null;
    } catch (_) { return null; }
  }

  /// Registers this device's FCM token with the backend. Safe to repeat; the
  /// backend is expected to upsert by token.
  Future<void> registerDeviceToken(String token) async {
    await dio.post('$apiBase/notifications/devices', data: {
      'token': token,
      'platform': 'android',
    });
  }

  Future<void> unregisterDeviceToken(String token) async {
    await dio.delete('$apiBase/notifications/devices', data: {'token': token});
  }

  Future<NotificationPreferences> fetchNotificationPreferences() async {
    final r = await dio.get('$apiBase/notifications/preferences');
    if (r.data is Map) {
      return NotificationPreferences.fromJson(
        Map<String, dynamic>.from(r.data as Map),
      );
    }
    throw ApiException('Could not load notification settings');
  }

  Future<NotificationPreferences> updateNotificationPreferences(
    Map<String, bool> changes,
  ) async {
    final r = await dio.put('$apiBase/notifications/preferences', data: changes);
    if (r.data is Map) {
      return NotificationPreferences.fromJson(
        Map<String, dynamic>.from(r.data as Map),
      );
    }
    throw ApiException('Could not save notification settings');
  }
}
