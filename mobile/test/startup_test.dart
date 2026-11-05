import 'dart:io';

import 'package:cookie_jar/cookie_jar.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:sandwitches/core/startup.dart';
import 'package:sandwitches/core/storage.dart';

class _FakeStorage extends AppStorage {
  _FakeStorage(this.baseUrl);

  final String? baseUrl;

  @override
  Future<String?> getBaseUrl() async => baseUrl;
}

void main() {
  const base = 'https://sandwitches.example';
  late Directory dir;
  late PersistCookieJar jar;

  setUp(() async {
    dir = await Directory.systemTemp.createTemp('sandwitches_startup_test_');
    jar = PersistCookieJar(storage: FileStorage(dir.path));
  });

  tearDown(() async {
    if (await dir.exists()) {
      await dir.delete(recursive: true);
    }
  });

  Future<void> seedCookie() async {
    await jar.saveFromResponse(Uri.parse(base), [
      Cookie(sessionCookieName, 'abc'),
    ]);
  }

  test(
    'no saved URL boots to onboarding without checking the session',
    () async {
      var called = false;

      final result = await resolveStartupDestination(
        storage: _FakeStorage(null),
        cookieJar: jar,
        checkSession: (baseUrl) async {
          called = true;
          return true;
        },
      );

      expect(result.destination, StartupDestination.onboarding);
      expect(result.baseUrl, isNull);
      expect(called, isFalse);
    },
  );

  test('offline with a stored cookie boots to home', () async {
    await seedCookie();

    final result = await resolveStartupDestination(
      storage: _FakeStorage(base),
      cookieJar: jar,
      checkSession: (baseUrl) async => null,
    );

    expect(result.destination, StartupDestination.home);
    expect(result.baseUrl, base);
  });

  test('offline without a cookie boots to login', () async {
    final result = await resolveStartupDestination(
      storage: _FakeStorage(base),
      cookieJar: jar,
      checkSession: (baseUrl) async => null,
    );

    expect(result.destination, StartupDestination.login);
    expect(result.baseUrl, base);
  });

  test('server accepts the session and boots to home', () async {
    final result = await resolveStartupDestination(
      storage: _FakeStorage(base),
      cookieJar: jar,
      checkSession: (baseUrl) async => true,
    );

    expect(result.destination, StartupDestination.home);
    expect(result.baseUrl, base);
  });

  test(
    'server rejects the session, boots to login and drops the cookie',
    () async {
      await seedCookie();

      final result = await resolveStartupDestination(
        storage: _FakeStorage(base),
        cookieJar: jar,
        checkSession: (baseUrl) async => false,
      );

      expect(result.destination, StartupDestination.login);
      expect(result.baseUrl, base);
      expect(await jar.loadForRequest(Uri.parse(base)), isEmpty);
    },
  );
}
