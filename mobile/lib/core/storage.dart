import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class AppStorage {
  static const _kBaseUrl = 'base_url';
  static const _kTheme = 'theme_mode';
  static const _kCookies = 'cookies';
  final FlutterSecureStorage _s = const FlutterSecureStorage();
  Future<String?> getBaseUrl() => _s.read(key: _kBaseUrl);
  Future<void> setBaseUrl(String v) => _s.write(key: _kBaseUrl, value: v);
  Future<String?> getTheme() => _s.read(key: _kTheme);
  Future<void> setTheme(String v) => _s.write(key: _kTheme, value: v);
  Future<String?> getCookies() => _s.read(key: _kCookies);
  Future<void> setCookies(String v) => _s.write(key: _kCookies, value: v);
  Future<void> clearAll() => _s.deleteAll();
  Future<void> clearAuth() async {
    await _s.delete(key: _kCookies);
  }
}
