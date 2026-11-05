import 'package:cookie_jar/cookie_jar.dart';

import 'storage.dart';

enum StartupDestination { onboarding, login, home }

class StartupResult {
  const StartupResult({required this.baseUrl, required this.destination});

  final String? baseUrl;
  final StartupDestination destination;
}

const String sessionCookieName = 'sandwitches_session';

/// Decides where the app should boot based on the saved server URL and the
/// persisted session cookie. When the server cannot be reached, a stored
/// session cookie is treated as a valid session so the app still opens offline.
Future<StartupResult> resolveStartupDestination({
  required AppStorage storage,
  required PersistCookieJar cookieJar,
  required Future<bool?> Function(String baseUrl) checkSession,
}) async {
  final raw = await storage.getBaseUrl();
  final trimmed = raw?.trim();
  if (trimmed == null || trimmed.isEmpty) {
    return const StartupResult(
      baseUrl: null,
      destination: StartupDestination.onboarding,
    );
  }

  final baseUrl = trimmed.endsWith('/')
      ? trimmed.substring(0, trimmed.length - 1)
      : trimmed;

  var hasSessionCookie = false;
  try {
    await cookieJar.forceInit();
    final cookies = await cookieJar.loadForRequest(Uri.parse(baseUrl));
    hasSessionCookie = cookies.any(
      (c) => c.name == sessionCookieName && c.value.isNotEmpty,
    );
  } catch (_) {
    hasSessionCookie = false;
  }

  bool? sessionValid;
  try {
    sessionValid = await checkSession(baseUrl);
  } catch (_) {
    sessionValid = null;
  }

  if (sessionValid == true || (sessionValid == null && hasSessionCookie)) {
    return StartupResult(
      baseUrl: baseUrl,
      destination: StartupDestination.home,
    );
  }

  if (sessionValid == false) {
    // Stale/revoked session: drop the cookie so the next launch is clean.
    try {
      await cookieJar.deleteAll();
    } catch (_) {}
  }

  return StartupResult(baseUrl: baseUrl, destination: StartupDestination.login);
}
