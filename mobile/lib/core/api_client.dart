import 'package:dio/dio.dart';
import 'package:dio_cookie_manager/dio_cookie_manager.dart';
import 'package:cookie_jar/cookie_jar.dart';

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
}
