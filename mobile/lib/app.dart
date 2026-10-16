import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'router/app_router.dart';
import 'theme/app_theme.dart';
import 'theme/theme_provider.dart';
import 'core/providers.dart';

class SandwitchesApp extends ConsumerStatefulWidget {
  const SandwitchesApp({super.key});
  @override
  ConsumerState<SandwitchesApp> createState()=> _AppState();
}
class _AppState extends ConsumerState<SandwitchesApp> {
  @override
  void initState(){
    super.initState();
    Future.microtask(() async {
      final s = ref.read(storageProvider);
      final base = await s.getBaseUrl();
      if (base!=null) {
        ref.read(baseUrlProvider.notifier).state = base;
      }
      final theme = await s.getTheme();
      if (theme!=null) {
        await ref.read(themeNotifierProvider.notifier).setTheme(theme);
      }
    });
  }
  @override
  Widget build(BuildContext context){
    final router = ref.watch(appRouterProvider);
    final themeMode = ref.watch(themeNotifierProvider);
    late final ThemeData theme;
    late final ThemeData darkTheme;
    if (themeMode=='eink') {
      theme = AppTheme.eink();
      darkTheme = AppTheme.eink();
    } else if (themeMode=='light') {
      theme = AppTheme.light();
      darkTheme = AppTheme.dark();
    } else if (themeMode=='dark') {
      theme = AppTheme.light();
      darkTheme = AppTheme.dark();
    } else {
      theme = AppTheme.light();
      darkTheme = AppTheme.dark();
    }
    late final ThemeMode mode;
    if (themeMode=='light') {
      mode=ThemeMode.light;
    } else if (themeMode=='dark') {
      mode=ThemeMode.dark;
    } else if (themeMode=='eink') {
      mode=ThemeMode.light;
    } else {
      mode=ThemeMode.system;
    }
    if (themeMode=='eink') {
      return MaterialApp.router(routerConfig: router, theme: AppTheme.eink(), darkTheme: AppTheme.eink(), themeMode: ThemeMode.light, title: 'Sandwitches');
    }
    return MaterialApp.router(routerConfig: router, theme: theme, darkTheme: darkTheme, themeMode: mode, title: 'Sandwitches');
  }
}
