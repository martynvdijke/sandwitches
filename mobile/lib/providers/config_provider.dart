import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

class ConfigProvider with ChangeNotifier {
  static const String _urlKey = 'sandwitches_api_url';
  String? _apiUrl;
  bool _isInitialized = false;

  String? get apiUrl => _apiUrl;
  bool get isInitialized => _isInitialized;

  Future<void> init() async {
    final prefs = await SharedPreferences.getInstance();
    _apiUrl = prefs.getString(_urlKey);
    _isInitialized = true;
    notifyListeners();
  }

  Future<void> saveApiUrl(String url) async {
    // Normalize the URL
    String normalizedUrl = url.trim();
    if (!normalizedUrl.endsWith('/')) {
      normalizedUrl += '/';
    }
    // If user didn't include 'api/v1' or similar, we should ensure the trailing structure.
    // However, the ApiService already appends 'v1/recipes'.
    // So the base URL should probably be the instance root + 'api/'.

    if (!normalizedUrl.endsWith('api/')) {
      // Check if it ends with 'api'
      if (normalizedUrl.endsWith('api')) {
        normalizedUrl += '/';
      } else {
        normalizedUrl += 'api/';
      }
    }

    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_urlKey, normalizedUrl);
    _apiUrl = normalizedUrl;
    notifyListeners();
  }

  Future<void> clear() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_urlKey);
    _apiUrl = null;
    notifyListeners();
  }
}
