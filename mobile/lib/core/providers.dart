import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/legacy.dart';
import 'package:dio/dio.dart';
import 'package:cookie_jar/cookie_jar.dart';
import 'package:dio_cookie_manager/dio_cookie_manager.dart';

import 'config.dart';
import 'storage.dart';
import 'models.dart';

final storageProvider = Provider<AppStorage>((_) => AppStorage());

final baseUrlProvider = StateProvider<String?>((_) => null);

final themeModeProvider = StateProvider<String>(
  (_) => 'system',
); // light,dark,eink,system

final persistCookieJarProvider = Provider<PersistCookieJar>(
  (_) => PersistCookieJar(),
);

final dioProvider = Provider<Dio>((ref) {
  final baseUrl = ref.watch(baseUrlProvider) ?? kDefaultBaseUrl;
  final jar = ref.watch(persistCookieJarProvider);
  final dio = Dio(
    BaseOptions(
      baseUrl: baseUrl.endsWith('/')
          ? baseUrl.substring(0, baseUrl.length - 1)
          : baseUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 10),
      followRedirects: false,
      validateStatus: (s) => s != null && s < 500,
    ),
  );
  dio.interceptors.add(CookieManager(jar));
  return dio;
});

final apiBaseProvider = Provider<String>((ref) {
  final b = ref.watch(baseUrlProvider) ?? kDefaultBaseUrl;
  final base = b.endsWith('/') ? b.substring(0, b.length - 1) : b;
  return '$base/api/v1';
});

// Repository
class DemoLoginConfig {
  const DemoLoginConfig({
    required this.enabled,
    required this.username,
    required this.password,
  });

  const DemoLoginConfig.disabled()
    : enabled = false,
      username = '',
      password = '';

  final bool enabled;
  final String username;
  final String password;

  factory DemoLoginConfig.fromJson(Map<String, dynamic> json) {
    final demo = json['demo'];
    if (demo is! Map) return const DemoLoginConfig.disabled();
    final enabled = demo['enabled'];
    return DemoLoginConfig(
      enabled: enabled is bool && enabled,
      username: demo['username'] is String ? demo['username'] as String : '',
      password: demo['password'] is String ? demo['password'] as String : '',
    );
  }
}

class ApiException implements Exception {
  final String message;
  ApiException(this.message);
  @override
  String toString() => message;
}

class ApiRepository {
  final Dio dio;
  final String baseUrl;
  ApiRepository(this.dio, this.baseUrl);
  String get apiBase {
    final b = baseUrl.endsWith('/')
        ? baseUrl.substring(0, baseUrl.length - 1)
        : baseUrl;
    return '$b/api/v1';
  }

  Future<bool> ping(String base) async {
    try {
      final b = base.endsWith('/') ? base.substring(0, base.length - 1) : base;
      final r = await dio.get('$b/api/v1/ping');
      return r.statusCode != null && r.statusCode! < 400;
    } catch (_) {
      return false;
    }
  }

  Future<List<Recipe>> fetchRecipes({
    String? search,
    String? tag,
    String? isApproved,
    int? page,
    int? perPage,
  }) async {
    final qp = <String, dynamic>{};
    if (search != null && search.isNotEmpty) qp['search'] = search;
    if (tag != null && tag.isNotEmpty) qp['tag'] = tag;
    if (isApproved != null) qp['is_approved'] = isApproved;
    if (page != null) qp['page'] = page;
    if (perPage != null) qp['per_page'] = perPage;
    final r = await dio.get('$apiBase/recipes', queryParameters: qp);
    final d = r.data;
    if (d is List) {
      return d.map((e) => Recipe.fromJson(e as Map<String, dynamic>)).toList();
    }
    if (d is Map && d['items'] is List) {
      return (d['items'] as List)
          .map((e) => Recipe.fromJson(e as Map<String, dynamic>))
          .toList();
    }
    if (d is Map && d['results'] is List) {
      return (d['results'] as List)
          .map((e) => Recipe.fromJson(e as Map<String, dynamic>))
          .toList();
    }
    return [];
  }

  Future<Recipe?> fetchRecipeOfDay() async {
    try {
      final r = await dio.get('$apiBase/recipe-of-the-day');
      if (r.statusCode == 404 || r.data == null) return null;
      if (r.data is Map<String, dynamic> && (r.data as Map).isEmpty) {
        return null;
      }
      if (r.data is Map<String, dynamic>) {
        return Recipe.fromJson(r.data as Map<String, dynamic>);
      }
      return null;
    } catch (_) {
      return null;
    }
  }

  Future<List<Tag>> fetchTags() async {
    final r = await dio.get('$apiBase/tags');
    if (r.data is List) {
      return (r.data as List)
          .map((e) => Tag.fromJson(e as Map<String, dynamic>))
          .toList();
    }
    return [];
  }

  Future<Recipe> fetchRecipe(int id) async {
    final r = await dio.get('$apiBase/recipes/$id');
    return Recipe.fromJson(r.data as Map<String, dynamic>);
  }

  Future<Recipe> createRecipe({
    required String title,
    String description = '',
    String ingredients = '',
    String instructions = '',
    int servings = 1,
    double? price,
    List<String> tags = const [],
    String? imagePath,
  }) async {
    final form = FormData.fromMap({
      'title': title,
      if (description.isNotEmpty) 'description': description,
      if (ingredients.isNotEmpty) 'ingredients': ingredients,
      if (instructions.isNotEmpty) 'instructions': instructions,
      'servings': servings,
      if (price != null) 'price': price.toString(),
      if (tags.isNotEmpty) 'tags': tags.join(','),
      if (imagePath != null) 'image': await MultipartFile.fromFile(imagePath),
    });
    final r = await dio.post(
      '$apiBase/recipes',
      data: form,
      options: Options(
        sendTimeout: const Duration(seconds: 60),
        receiveTimeout: const Duration(seconds: 60),
      ),
    );
    if (r.statusCode == 201 && r.data is Map<String, dynamic>) {
      return Recipe.fromJson(r.data as Map<String, dynamic>);
    }
    final data = r.data;
    final message = data is Map && data['message'] != null
        ? data['message'].toString()
        : 'Could not create recipe (${r.statusCode})';
    throw ApiException(message);
  }

  Future<List<ScaledIngredient>> scaleIngredients(int id, int target) async {
    final r = await dio.get(
      '$apiBase/recipes/$id/scale-ingredients',
      queryParameters: {'target_servings': target},
    );
    if (r.data is List) {
      return (r.data as List)
          .map((e) => ScaledIngredient.fromJson(e as Map<String, dynamic>))
          .toList();
    }
    return [];
  }

  Future<Map<String, dynamic>> fetchRatingSummary(int id) async {
    final r = await dio.get('$apiBase/recipes/$id/rating');
    if (r.data is Map) return Map<String, dynamic>.from(r.data as Map);
    return {'average': 0, 'count': 0};
  }

  Future<List<Rating>> fetchRatings(int id) async {
    try {
      final r = await dio.get('$apiBase/recipes/$id/ratings');
      if (r.data is List) {
        return (r.data as List)
            .map((e) => Rating.fromJson(e as Map<String, dynamic>))
            .toList();
      }
      if (r.data is Map && r.data['items'] is List) {
        return (r.data['items'] as List)
            .map((e) => Rating.fromJson(e as Map<String, dynamic>))
            .toList();
      }
      return [];
    } catch (_) {
      return [];
    }
  }

  Future<Rating> postRating(
    int id,
    int score,
    String comment, {
    String? imagePath,
  }) async {
    final form = FormData.fromMap({
      'score': score.toString(),
      'comment': comment,
      if (imagePath != null) 'image': await MultipartFile.fromFile(imagePath),
    });
    final r = await dio.post('$apiBase/recipes/$id/ratings', data: form);
    return Rating.fromJson(r.data as Map<String, dynamic>);
  }

  Future<List<CartItem>> fetchCart() async {
    final r = await dio.get('$apiBase/cart');
    if (r.data is List) {
      return (r.data as List)
          .map((e) => CartItem.fromJson(e as Map<String, dynamic>))
          .toList();
    }
    return [];
  }

  Future<CartItem> addToCart(int recipeId, int qty) async {
    final r = await dio.post(
      '$apiBase/cart',
      data: {'recipe_id': recipeId, 'quantity': qty},
    );
    return CartItem.fromJson(r.data as Map<String, dynamic>);
  }

  Future<CartItem> updateCart(int id, int qty) async {
    final r = await dio.patch('$apiBase/cart/$id', data: {'quantity': qty});
    return CartItem.fromJson(r.data as Map<String, dynamic>);
  }

  Future<void> deleteCart(int id) async {
    await dio.delete('$apiBase/cart/$id');
  }

  Future<List<Order>> fetchOrders() async {
    final r = await dio.get('$apiBase/orders');
    if (r.data is List) {
      return (r.data as List)
          .map((e) => Order.fromJson(e as Map<String, dynamic>))
          .toList();
    }
    return [];
  }

  Future<Order> createOrder(int recipeId) async {
    final r = await dio.post('$apiBase/orders', data: {'recipe_id': recipeId});
    return Order.fromJson(r.data as Map<String, dynamic>);
  }

  Future<OrderDetail> fetchOrder(int id) async {
    final r = await dio.get('$apiBase/orders/$id');
    return OrderDetail.fromJson(r.data as Map<String, dynamic>);
  }

  Future<OrderDetail> trackOrder(String token) async {
    final r = await dio.get('$apiBase/orders/track/$token');
    return OrderDetail.fromJson(r.data as Map<String, dynamic>);
  }

  Future<void> patchOrderStatus(int id, String status) async {
    await dio.patch('$apiBase/orders/$id', data: {'status': status});
  }

  Future<Map<String, dynamic>?> fetchMe() async {
    try {
      final r = await dio.get('$apiBase/me');
      if (r.statusCode == 200 && r.data is Map) {
        return Map<String, dynamic>.from(r.data as Map);
      }
      return null;
    } catch (_) {
      return null;
    }
  }

  /// Returns true when the persisted session is accepted by the server,
  /// false when the server rejects it (401/403), and null when the server
  /// could not be reached or returned an unexpected status.
  Future<bool?> checkSession() async {
    try {
      final r = await dio.get('$apiBase/me');
      if (r.statusCode == 200) return true;
      if (r.statusCode == 401 || r.statusCode == 403) return false;
      return null;
    } catch (_) {
      return null;
    }
  }

  /// Fetches the public demo-login configuration. Any failure yields a
  /// disabled config so the login screen never breaks because of it.
  Future<DemoLoginConfig> fetchAuthConfig() async {
    try {
      final r = await dio.get('$apiBase/auth/config');
      if (r.statusCode == 200 && r.data is Map) {
        return DemoLoginConfig.fromJson(
          Map<String, dynamic>.from(r.data as Map),
        );
      }
      return const DemoLoginConfig.disabled();
    } catch (_) {
      return const DemoLoginConfig.disabled();
    }
  }

  /// Downloads the authenticated account export as raw JSON bytes.
  Future<Uint8List> exportAccount() async {
    try {
      final r = await dio.get(
        '$apiBase/me/export',
        options: Options(responseType: ResponseType.bytes),
      );
      if (r.statusCode == 200) {
        final data = r.data;
        if (data is Uint8List) return data;
        if (data is List<int>) return Uint8List.fromList(data);
        return Uint8List(0);
      }
      throw ApiException(_exportErrorMessage(r.statusCode, r.data));
    } on ApiException {
      rethrow;
    } on DioException catch (e) {
      throw ApiException(
        _exportErrorMessage(e.response?.statusCode, e.response?.data),
      );
    } catch (_) {
      throw ApiException('Could not download your data');
    }
  }

  String _exportErrorMessage(int? statusCode, dynamic data) {
    if (data is Map && data['message'] != null) {
      return data['message'].toString();
    }
    if (data is List<int>) {
      try {
        final decoded = jsonDecode(utf8.decode(data));
        if (decoded is Map && decoded['message'] != null) {
          return decoded['message'].toString();
        }
      } catch (_) {}
    }
    return 'Could not download your data (${statusCode ?? 'network error'})';
  }

  Future<void> deleteAccount(String confirm) async {
    final r = await dio.delete('$apiBase/me', data: {'confirm': confirm});
    if (r.statusCode == 204 || r.statusCode == 200) return;
    final data = r.data;
    final message = data is Map && data['message'] != null
        ? data['message'].toString()
        : 'Could not delete the account (${r.statusCode})';
    throw ApiException(message);
  }

  Future<void> registerDeviceToken(String token) async {
    await dio.post(
      '$apiBase/notifications/devices',
      data: {'token': token, 'platform': 'android'},
    );
  }

  Future<void> unregisterDeviceToken(String token) async {
    await dio.delete(
      '$apiBase/notifications/devices',
      data: {'token': token},
    );
  }

  Future<NotificationPreferences> fetchNotificationPreferences() async {
    final r = await dio.get('$apiBase/notifications/preferences');
    final data = r.data;
    if (data is Map) {
      return NotificationPreferences.fromJson(
        Map<String, dynamic>.from(data),
      );
    }
    throw ApiException(
      _notificationErrorMessage(
        r.statusCode,
        data,
        'Could not load notification settings',
      ),
    );
  }

  Future<NotificationPreferences> updateNotificationPreferences(
    Map<String, bool> changes,
  ) async {
    final r = await dio.put(
      '$apiBase/notifications/preferences',
      data: changes,
    );
    final data = r.data;
    if (data is Map) {
      return NotificationPreferences.fromJson(
        Map<String, dynamic>.from(data),
      );
    }
    throw ApiException(
      _notificationErrorMessage(
        r.statusCode,
        data,
        'Could not save notification settings',
      ),
    );
  }

  String _notificationErrorMessage(
    int? statusCode,
    dynamic data,
    String fallback,
  ) {
    if (data is Map && data['message'] != null) {
      return data['message'].toString();
    }
    return '$fallback (${statusCode ?? 'network error'})';
  }
}

final repositoryProvider = Provider<ApiRepository>((ref) {
  final dio = ref.watch(dioProvider);
  final base = ref.watch(baseUrlProvider) ?? kDefaultBaseUrl;
  return ApiRepository(dio, base);
});

final authConfigProvider = FutureProvider<DemoLoginConfig>(
  (ref) => ref.watch(repositoryProvider).fetchAuthConfig(),
);

// Future providers
final tagsProvider = FutureProvider<List<Tag>>(
  (ref) => ref.watch(repositoryProvider).fetchTags(),
);
final recipeOfDayProvider = FutureProvider<Recipe?>(
  (ref) => ref.watch(repositoryProvider).fetchRecipeOfDay(),
);
final recipesProvider =
    FutureProvider.family<List<Recipe>, ({String search, String tag})>(
      (ref, args) => ref
          .watch(repositoryProvider)
          .fetchRecipes(search: args.search, tag: args.tag),
    );
final meProvider = FutureProvider<Map<String, dynamic>?>(
  (ref) => ref.watch(repositoryProvider).fetchMe(),
);
final cartProvider = FutureProvider<List<CartItem>>(
  (ref) => ref.watch(repositoryProvider).fetchCart(),
);
final ordersProvider = FutureProvider<List<Order>>(
  (ref) => ref.watch(repositoryProvider).fetchOrders(),
);
