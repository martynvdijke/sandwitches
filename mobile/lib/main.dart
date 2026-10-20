import 'package:cookie_jar/cookie_jar.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:path_provider/path_provider.dart';
import 'app.dart';
import 'core/providers.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // PersistCookieJar defaults to a relative `.cookies` directory, which is
  // read-only on Android and makes every request fail while loading cookies.
  final supportDir = await getApplicationSupportDirectory();
  final cookieJar = PersistCookieJar(
    storage: FileStorage('${supportDir.path}/cookies'),
  );

  runApp(ProviderScope(
    overrides: [persistCookieJarProvider.overrideWithValue(cookieJar)],
    child: const SandwitchesApp(),
  ));
}
