// ignore_for_file: avoid_print
import 'dart:io';
import 'dart:typed_data';
import 'package:flutter_test/flutter_test.dart';
import 'package:dio/dio.dart';
import 'package:cookie_jar/cookie_jar.dart';
import 'package:dio_cookie_manager/dio_cookie_manager.dart';
import 'package:sandwitches/core/providers.dart';

void main() {
  final url = Platform.environment['SANDWITCHES_E2E_URL'];

  if (url == null || url.isEmpty) {
    test('live e2e (skipped without SANDWITCHES_E2E_URL)', () {
      print('SKIP: SANDWITCHES_E2E_URL not set');
    }, skip: 'Set SANDWITCHES_E2E_URL to run live e2e');
    return;
  }

  test('live e2e: app API layer against live Go backend', () async {
    final baseUrl = url.endsWith('/') ? url.substring(0, url.length - 1) : url;
    final tmpDir = await Directory.systemTemp.createTemp('sandwitches_e2e_cookies_');
    final pngFile = File('${tmpDir.path}/tiny.png');

    // 1x1 valid PNG
    final pngBytes = Uint8List.fromList([
      0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
      0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
      0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
      0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xDE,
      0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41, 0x54,
      0x08, 0xD7, 0x63, 0xF8, 0xFF, 0xFF, 0x3F, 0x00, 0x05, 0xFE, 0x02, 0xFE, 0xDC, 0xCC, 0x59, 0xE7,
      0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
    ]);
    await pngFile.writeAsBytes(pngBytes);

    final jar = PersistCookieJar(storage: FileStorage(tmpDir.path));
    final dio = Dio(BaseOptions(
      baseUrl: baseUrl,
      followRedirects: false,
      validateStatus: (s) => s != null && s < 500,
    ));
    dio.interceptors.add(CookieManager(jar));
    final repo = ApiRepository(dio, baseUrl);

    final results = <String, String>{};
    void pass(String step, String msg) {
      results[step] = 'PASS: $msg';
      print('PASS [$step] $msg');
    }

    // ignore: unused_element
    void fail(String step, String msg) {
      results[step] = 'FAIL: $msg';
      print('FAIL [$step] $msg');
    }

    try {
      // ping
      {
        final ok = await repo.ping(baseUrl);
        expect(ok, isTrue, reason: 'ping');
        pass('ping', 'GET /api/v1/ping ok');
      }

      // setup
      {
        final r = await dio.post('/setup',
            data: {
              'username': 'e2eadmin',
              'password1': 'TestPass123!',
              'password2': 'TestPass123!',
              'email': 'e2e@example.com',
            },
            options: Options(
              contentType: Headers.formUrlEncodedContentType,
              followRedirects: false,
              validateStatus: (s) => s != null && s < 500,
            ));
        // 302 on success or if already exists (superuser exists)
        expect(r.statusCode, anyOf([200, 302, 303]),
            reason: 'setup status ${r.statusCode} body=${r.data}');
        pass('setup', 'POST /setup ${r.statusCode}');
      }

      // login
      {
        final r = await dio.post('/login',
            data: {'username': 'e2eadmin', 'password': 'TestPass123!'},
            options: Options(
              contentType: Headers.formUrlEncodedContentType,
              followRedirects: false,
              validateStatus: (s) => s != null && s < 500,
            ));
        expect(r.statusCode, anyOf([200, 302, 303]),
            reason: 'login status ${r.statusCode}');
        pass('login', 'POST /login ${r.statusCode}');
      }

      // /me
      {
        final me = await repo.fetchMe();
        expect(me, isNotNull, reason: '/me should be 200 after login');
        pass('me', 'GET /api/v1/me 200 user=${me!['username'] ?? me['email'] ?? me}');
      }

      // create 2 recipes
      int id1, id2;
      {
        Future<int> create(String title, String tagName) async {
          final r = await dio.post('$baseUrl/api/v1/recipes', data: {
            'title': title,
            'description': 'desc $title',
            'ingredients': '2 cups flour\n1 cup sugar',
            'instructions': 'Mix and bake.',
            'servings': 4,
            'price': 9.99,
            'tags': [tagName],
          }, options: Options(contentType: Headers.jsonContentType));
          expect(r.statusCode, 201, reason: 'create $title ${r.data}');
          final j = r.data as Map<String, dynamic>;
          return (j['id'] as num).toInt();
        }

        id1 = await create('E2E Recipe One', 'E2E-Tag-A');
        id2 = await create('E2E Recipe Two', 'E2E-Tag-B');
        pass('create_recipes', 'created $id1 and $id2');
      }

      // list recipes contains them
      {
        final list = await repo.fetchRecipes();
        expect(list.any((r) => r.id == id1), isTrue);
        expect(list.any((r) => r.id == id2), isTrue);
        pass('list', 'GET /api/v1/recipes contains both (${list.length} total)');
      }

      // search filters correctly
      {
        final filtered = await repo.fetchRecipes(search: 'Recipe One');
        expect(filtered.any((r) => r.id == id1), isTrue);
        // should NOT contain recipe two (title doesn't match)
        expect(filtered.any((r) => r.id == id2), isFalse,
            reason: 'search should filter out non-matching');
        pass('search', 'search=Recipe One filtered correctly (${filtered.length})');
      }

      // tags list non-empty
      {
        final tags = await repo.fetchTags();
        expect(tags.isNotEmpty, isTrue);
        pass('tags', 'GET /api/v1/tags non-empty (${tags.length})');
      }

      // recipe detail
      {
        final recipe = await repo.fetchRecipe(id1);
        expect(recipe.id, id1);
        pass('detail', 'GET /api/v1/recipes/$id1 title=${recipe.title}');
      }

      // scaleIngredients
      {
        final rows = await repo.scaleIngredients(id1, 8);
        expect(rows.isNotEmpty, isTrue, reason: 'scaleIngredients should return rows');
        pass('scale', 'scale target 8 returned ${rows.length} rows');
      }

      // post rating JSON path
      {
        final r = await dio.post('$baseUrl/api/v1/recipes/$id1/ratings',
            data: {'score': 5, 'comment': 'json rating'},
            options: Options(contentType: Headers.jsonContentType));
        expect(r.statusCode, anyOf([200, 201]),
            reason: 'post rating json ${r.statusCode} ${r.data}');
        pass('rating_json', 'POST rating JSON  score=5');
      }

      // post rating WITH image (multipart via repo)
      {
        final rating = await repo.postRating(id1, 4, 'with image', imagePath: pngFile.path);
        expect(rating.image, isNotNull, reason: 'image field should be non-null');
        expect(rating.image!.isNotEmpty, isTrue);
        pass('rating_image', 'POST rating multipart image=${rating.image}');
      }

      // GET ratings list (backend upserts per user: 2nd POST updates 1st, so expect 1)
      {
        final ratings = await repo.fetchRatings(id1);
        expect(ratings.length, greaterThanOrEqualTo(1));
        // after multipart update, image should be present
        expect(ratings.first.image, isNotNull);
        expect(ratings.first.image!.isNotEmpty, isTrue);
        pass('ratings_list', 'GET ratings list ${ratings.length} entries image=${ratings.first.image}');
      }

      // cart add/update/delete
      {
        final added = await repo.addToCart(id1, 2);
        expect(added.quantity, 2);
        pass('cart_add', 'POST /api/v1/cart id=${added.id} qty=2');

        final updated = await repo.updateCart(added.id, 3);
        expect(updated.quantity, 3);
        pass('cart_update', 'PATCH /api/v1/cart/${added.id} qty=3');

        final cart = await repo.fetchCart();
        expect(cart.any((c) => c.id == added.id), isTrue);
        pass('cart_list', 'GET /api/v1/cart ${cart.length} items');

        await repo.deleteCart(added.id);
        final after = await repo.fetchCart();
        expect(after.any((c) => c.id == added.id), isFalse);
        pass('cart_delete', 'DELETE /api/v1/cart/${added.id} removed');
      }

      // create order
      late String trackingToken;
      late int orderId;
      {
        final order = await repo.createOrder(id1);
        expect(order.trackingToken.isNotEmpty, isTrue,
            reason: 'trackingToken non-empty ${order.trackingToken}');
        trackingToken = order.trackingToken;
        orderId = order.id;
        pass('create_order', 'POST /api/v1/orders id=$orderId token=$trackingToken');
      }

      // orders list/detail
      {
        final orders = await repo.fetchOrders();
        expect(orders.any((o) => o.id == orderId), isTrue);
        pass('orders_list', 'GET /api/v1/orders contains $orderId');

        final detail = await repo.fetchOrder(orderId);
        expect(detail.items.isNotEmpty, isTrue);
        pass('order_detail', 'GET /api/v1/orders/$orderId items=${detail.items.length}');
      }

      // public trackOrder
      {
        final tracked = await repo.trackOrder(trackingToken);
        expect(tracked.id, orderId);
        pass('track', 'GET /api/v1/orders/track/$trackingToken id=${tracked.id}');
      }

      // staff patchOrderStatus
      {
        await repo.patchOrderStatus(orderId, 'MAKING');
        final refetched = await repo.fetchOrder(orderId);
        expect(refetched.status, 'MAKING', reason: 'status should be MAKING got ${refetched.status}');
        pass('patch_status', 'PATCH status MAKING verified');
      }

      print('\n=== E2E SUMMARY ===');
      results.forEach((k, v) => print('  $k: $v'));
      print('=== ALL STEPS PASSED ===\n');
    } finally {
      try {
        await tmpDir.delete(recursive: true);
      } catch (_) {}
    }
  });
}
