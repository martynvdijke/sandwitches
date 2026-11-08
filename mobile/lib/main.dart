import 'dart:async';

import 'package:cookie_jar/cookie_jar.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path_provider/path_provider.dart';
import 'app.dart';
import 'core/config.dart';
import 'core/providers.dart';
import 'core/storage.dart';
import 'src/telemetry/telemetry.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // PersistCookieJar defaults to a relative `.cookies` directory, which is
  // read-only on Android and makes every request fail while loading cookies.
  final supportDir = await getApplicationSupportDirectory();
  final cookieJar = PersistCookieJar(
    storage: FileStorage('${supportDir.path}/cookies'),
  );

  // Resolve the same persisted server URL the app uses so telemetry reports to
  // the instance the user actually talks to. Never blocks the first frame.
  unawaited(() async {
    String baseUrl;
    try {
      baseUrl = await AppStorage().getBaseUrl() ?? kDefaultBaseUrl;
    } catch (_) {
      baseUrl = kDefaultBaseUrl;
    }
    await Telemetry.instance.startFromServer(baseUrl);
  }());

  runApp(ProviderScope(
    overrides: [persistCookieJarProvider.overrideWithValue(cookieJar)],
    child: const SandwitchesApp(),
  ));
}
