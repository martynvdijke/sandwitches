import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'providers/recipe_provider.dart';
import 'screens/recipe_list_screen.dart';

void main() {
  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => RecipeProvider()),
      ],
      child: const SandwitchesApp(),
    ),
  );
}

class SandwitchesApp extends StatelessWidget {
  const SandwitchesApp({super.key});

  @override
  Widget build(BuildContext context) {
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
      home: const RecipeListScreen(),
      debugShowCheckedModeBanner: false,
    );
  }
}
