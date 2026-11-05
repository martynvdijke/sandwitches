import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../core/providers.dart';
import '../../core/startup.dart';

class StartupScreen extends ConsumerStatefulWidget {
  const StartupScreen({super.key});

  @override
  ConsumerState<StartupScreen> createState() => _StartupState();
}

class _StartupState extends ConsumerState<StartupScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _bootstrap());
  }

  Future<void> _bootstrap() async {
    final result = await resolveStartupDestination(
      storage: ref.read(storageProvider),
      cookieJar: ref.read(persistCookieJarProvider),
      checkSession: (baseUrl) {
        ref.read(baseUrlProvider.notifier).state = baseUrl;
        return ref.read(repositoryProvider).checkSession();
      },
    );
    if (!mounted) return;
    if (result.baseUrl != null) {
      ref.read(baseUrlProvider.notifier).state = result.baseUrl;
    }
    switch (result.destination) {
      case StartupDestination.onboarding:
        context.go('/onboarding');
      case StartupDestination.login:
        context.go('/login');
      case StartupDestination.home:
        context.go('/');
    }
  }

  @override
  Widget build(BuildContext context) {
    return const Scaffold(body: Center(child: CircularProgressIndicator()));
  }
}
