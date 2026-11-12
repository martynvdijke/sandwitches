import 'dart:async';

import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';
import 'package:go_router/go_router.dart';

import 'providers.dart';

/// Push notifications for the Android client.
///
/// Firebase is optional: a build without `google-services.json` has no Firebase
/// options, [Firebase.initializeApp] throws, and every push feature here turns
/// itself off while the rest of the app boots normally.
///
/// The service is a process-wide singleton and is safe to initialize once.
class NotificationService {
  NotificationService._();

  static final NotificationService instance = NotificationService._();

  /// Channel id shared with the backend's default notification channel
  /// (`com.google.firebase.messaging.default_notification_channel_id`).
  static const String channelId = 'sandwitches_default';
  static const String _channelName = 'Sandwitches';
  static const String _channelDescription =
      'New recipes, order updates and account notices';

  final FlutterLocalNotificationsPlugin _local =
      FlutterLocalNotificationsPlugin();

  ApiRepository Function()? _repository;
  GoRouter? _router;

  Future<void>? _initializing;
  bool _localReady = false;
  bool _firebaseReady = false;
  bool _signedIn = false;

  String? _token;
  Map<String, dynamic>? _pendingData;

  StreamSubscription<RemoteMessage>? _onMessageSub;
  StreamSubscription<RemoteMessage>? _onOpenSub;
  StreamSubscription<String>? _tokenSub;

  bool get _isAndroid =>
      !kIsWeb && defaultTargetPlatform == TargetPlatform.android;

  /// Initializes local notifications and, when Firebase is configured, the
  /// messaging listeners. Never throws: failures are debug-logged so startup
  /// is unaffected.
  ///
  /// [repository] is read lazily because the API client is rebuilt whenever
  /// the configured server URL changes.
  Future<void> initialize({
    required ApiRepository Function() repository,
    GoRouter? router,
  }) {
    _repository = repository;
    if (router != null) _router = router;
    return _initializing ??= _init();
  }

  Future<void> _init() async {
    await _setupLocalNotifications();
    await _setupFirebase();
  }

  Future<void> _setupLocalNotifications() async {
    if (!_isAndroid) return;
    try {
      const settings = InitializationSettings(
        // A drawable is required (mipmap resources are rejected). The existing
        // launcher foreground is the closest match in this project.
        android: AndroidInitializationSettings('ic_launcher_foreground'),
      );
      await _local.initialize(
        settings: settings,
        onDidReceiveNotificationResponse: (response) {
          _handleOpenUrl(<String, dynamic>{'url': response.payload});
        },
      );
      final android = _local
          .resolvePlatformSpecificImplementation<
            AndroidFlutterLocalNotificationsPlugin
          >();
      await android?.createNotificationChannel(
        const AndroidNotificationChannel(
          channelId,
          _channelName,
          description: _channelDescription,
          importance: Importance.high,
        ),
      );
      _localReady = true;
    } catch (e) {
      debugPrint('Local notifications are unavailable: $e');
    }
  }

  Future<void> _setupFirebase() async {
    try {
      if (Firebase.apps.isEmpty) {
        await Firebase.initializeApp();
      }
      _firebaseReady = true;
    } catch (e) {
      // Most commonly: no google-services.json in this build. Push stays off.
      debugPrint('Firebase is not configured; push notifications are off: $e');
      return;
    }

    try {
      _onMessageSub ??= FirebaseMessaging.onMessage.listen(_showForeground);
      _onOpenSub ??= FirebaseMessaging.onMessageOpenedApp.listen((message) {
        _handleOpenUrl(message.data);
      });

      final initial = await FirebaseMessaging.instance.getInitialMessage();
      if (initial != null) {
        // The app may still be on the startup/login screens, so hold the
        // destination until a session is known to exist.
        _pendingData = initial.data;
      }

      _tokenSub ??= FirebaseMessaging.instance
          // ignore: deprecated_member_use
          .onTokenRefresh
          .listen((token) {
            _token = token;
            if (_signedIn) unawaited(_register(token));
          });
    } catch (e) {
      debugPrint('FCM listeners could not be registered: $e');
    }
  }

  /// Registers this device's FCM token with the backend. Call after a session
  /// is known to be valid (successful login or restored session). Safe to call
  /// more than once.
  Future<void> syncToken() async {
    _signedIn = true;
    _flushPendingNavigation();

    final initializing = _initializing;
    if (initializing != null) {
      try {
        await initializing;
      } catch (_) {}
    }
    if (!_firebaseReady) return;

    try {
      final settings = await FirebaseMessaging.instance.requestPermission();
      if (settings.authorizationStatus == AuthorizationStatus.denied) {
        return;
      }
    } catch (e) {
      debugPrint('Notification permission could not be requested: $e');
    }

    try {
      final android = _local
          .resolvePlatformSpecificImplementation<
            AndroidFlutterLocalNotificationsPlugin
          >();
      await android?.requestNotificationsPermission();
    } catch (e) {
      debugPrint('Android notification permission request failed: $e');
    }

    try {
      // ignore: deprecated_member_use
      final token = await FirebaseMessaging.instance.getToken();
      if (token == null || token.isEmpty) return;
      _token = token;
      await _register(token);
    } catch (e) {
      debugPrint('FCM token registration failed: $e');
    }
  }

  Future<void> _register(String token) async {
    final repository = _repository?.call();
    if (repository == null) return;
    try {
      await repository.registerDeviceToken(token);
    } catch (e) {
      debugPrint('Could not register the device for push: $e');
    }
  }

  /// Stops push for this device. Call before clearing the session so the
  /// backend still accepts the authenticated DELETE.
  Future<void> unregisterToken() async {
    _signedIn = false;
    final token = _token;
    _token = null;
    final repository = _repository?.call();
    if (repository != null && token != null) {
      try {
        await repository.unregisterDeviceToken(token);
      } catch (e) {
        debugPrint('Could not unregister the device from push: $e');
      }
    }
    if (_firebaseReady) {
      try {
        // ignore: deprecated_member_use
        await FirebaseMessaging.instance.deleteToken();
      } catch (e) {
        debugPrint('Could not delete the FCM token: $e');
      }
    }
  }

  Future<void> _showForeground(RemoteMessage message) async {
    if (!_localReady) return;
    final notification = message.notification;
    final title =
        notification?.title ?? message.data['title']?.toString() ?? 'Sandwitches';
    final body = notification?.body ?? message.data['body']?.toString() ?? '';
    if (title.isEmpty && body.isEmpty) return;
    try {
      await _local.show(
        id: (message.messageId ?? '$title$body').hashCode & 0x7fffffff,
        title: title,
        body: body,
        payload: message.data['url']?.toString(),
        notificationDetails: const NotificationDetails(
          android: AndroidNotificationDetails(
            channelId,
            _channelName,
            channelDescription: _channelDescription,
            importance: Importance.high,
            priority: Priority.high,
          ),
        ),
      );
    } catch (e) {
      debugPrint('Could not show the foreground notification: $e');
    }
  }

  void _handleOpenUrl(Map<String, dynamic>? data) {
    if (!_signedIn) {
      _pendingData = data ?? _pendingData;
      _flushPendingNavigation();
      return;
    }
    _goTo(_routeForPayload(data));
  }

  void _flushPendingNavigation() {
    if (!_signedIn || _pendingData == null) return;
    final data = _pendingData;
    _pendingData = null;
    _goTo(_routeForPayload(data));
  }

  void _goTo(String route) {
    final router = _router;
    if (router == null) return;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      try {
        router.go(route);
      } catch (e) {
        debugPrint('Could not open notification destination $route: $e');
      }
    });
  }

  /// Maps a backend/web URL from the push payload onto an app route. Unknown
  /// paths fall back to the home screen instead of failing the tap.
  static String _routeForPayload(Map<String, dynamic>? data) {
    final raw = data?['url']?.toString().trim();
    final recipeId = int.tryParse(data?['recipe_id']?.toString() ?? '');
    if (raw == null || raw.isEmpty) return '/';

    Uri? uri;
    try {
      uri = Uri.parse(raw);
    } catch (_) {
      return '/';
    }

    final segments = uri.path.split('/').where((s) => s.isNotEmpty).toList();
    if (segments.isEmpty) return '/';

    switch (segments.first) {
      case 'recipe':
      case 'recipes':
        final id = recipeId ??
            (segments.length >= 2 ? int.tryParse(segments[1]) : null);
        if (id == null) return '/';
        if (segments.length >= 3 && segments[2] == 'cooking') {
          return '/recipe/$id/cooking';
        }
        return '/recipe/$id';
      case 'orders':
      case 'order':
        if (segments.length >= 3 && segments[1] == 'track') {
          return '/orders/track/${segments[2]}';
        }
        return '/orders';
      case 'dashboard':
        if (segments.length >= 2 && segments[1] == 'orders') return '/orders';
        return '/';
      case 'cart':
        return '/cart';
      case 'admin':
        return '/admin';
      case 'settings':
        return '/settings';
      default:
        return '/';
    }
  }
}
