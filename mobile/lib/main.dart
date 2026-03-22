import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'providers/config_provider.dart';
import 'providers/recipe_provider.dart';
import 'screens/recipe_list_screen.dart';
import 'screens/setup_screen.dart';
import 'services/api_service.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final configProvider = ConfigProvider();
  await configProvider.init();

  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider.value(value: configProvider),
        ChangeNotifierProxyProvider<ConfigProvider, RecipeProvider?>(
          create: (context) => null,
          update: (context, config, previous) {
            if (config.apiUrl == null) return null;
            if (previous != null && previous.apiService.baseUrl == config.apiUrl) return previous;
            return RecipeProvider(
              apiService: ApiService(baseUrl: config.apiUrl!),
            );
          },
        ),
      ],
      child: const SandwitchesApp(),
    ),
  );
}

class SandwitchesApp extends StatelessWidget {
  const SandwitchesApp({super.key});

  @override
  Widget build(BuildContext context) {
    final config = context.watch<ConfigProvider>();

    return MaterialApp(
      title: 'Sandwitches',
      theme: ThemeData(
        useMaterial3: true,
        colorScheme: ColorScheme.fromSeed(
          seedColor: Colors.orange,
          primary: Colors.orange[800],
        ),
        appBarTheme: AppBarTheme(
          backgroundColor: Colors.orange[800],
          foregroundColor: Colors.white,
        ),
      ),
      home: config.apiUrl == null ? const SetupScreen() : const RecipeListScreen(),
      debugShowCheckedModeBanner: false,
    );
  }
}
