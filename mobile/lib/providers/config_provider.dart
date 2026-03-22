import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

class ConfigProvider with ChangeNotifier {
  static const String _urlKey = 'sandwitches_base_url';
  String? _baseUrl;
  bool _isInitialized = false;

  String? get baseUrl => _baseUrl;
  String? get apiUrl => _baseUrl != null ? '${_baseUrl}api/' : null;
  bool get isInitialized => _isInitialized;

  Future<void> init() async {
    final prefs = await SharedPreferences.getInstance();
    _baseUrl = prefs.getString(_urlKey);
    _isInitialized = true;
    notifyListeners();
  }

  Future<void> saveApiUrl(String url) async {
    // Normalize the URL
    String normalizedUrl = url.trim();
    if (!normalizedUrl.endsWith('/')) {
      normalizedUrl += '/';
    }

    // Remove 'api/' if user included it, so we have the root
    if (normalizedUrl.endsWith('api/')) {
      normalizedUrl = normalizedUrl.substring(0, normalizedUrl.length - 4);
    }

    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_urlKey, normalizedUrl);
    _baseUrl = normalizedUrl;
    notifyListeners();
  }

  Future<void> clear() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_urlKey);
    _baseUrl = null;
    notifyListeners();
  }
}
