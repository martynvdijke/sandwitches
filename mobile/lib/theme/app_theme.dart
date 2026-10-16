import 'package:flutter/material.dart';

class AppTheme {
  static ThemeData light() {
    return ThemeData(
      useMaterial3: true,
      colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepOrange, brightness: Brightness.light),
    );
  }
  static ThemeData dark() {
    return ThemeData(
      useMaterial3: true,
      colorScheme: ColorScheme.fromSeed(seedColor: Colors.deepOrange, brightness: Brightness.dark),
    );
  }
  static ThemeData eink() {
    final base = ThemeData(useMaterial3: true, brightness: Brightness.light);
    return base.copyWith(
      colorScheme: const ColorScheme.light(
        primary: Colors.black, onPrimary: Colors.white,
        secondary: Colors.black, onSecondary: Colors.white,
        surface: Colors.white, onSurface: Colors.black,
        error: Colors.black, onError: Colors.white,
        outline: Colors.black,
      ),
      scaffoldBackgroundColor: Colors.white,
      appBarTheme: const AppBarTheme(backgroundColor: Colors.white, foregroundColor: Colors.black, elevation: 0, surfaceTintColor: Colors.white),
      cardTheme: CardThemeData(
        color: Colors.white,
        elevation: 0,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(4), side: const BorderSide(color: Colors.black, width: 2)),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(style: OutlinedButton.styleFrom(side: const BorderSide(color: Colors.black, width:2), foregroundColor: Colors.black, minimumSize: const Size(48,48))),
      elevatedButtonTheme: ElevatedButtonThemeData(style: ElevatedButton.styleFrom(backgroundColor: Colors.black, foregroundColor: Colors.white, minimumSize: const Size(48,48), shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(4)))),
      chipTheme: const ChipThemeData(side: BorderSide(color: Colors.black, width:1.5)),
      inputDecorationTheme: const InputDecorationTheme(border: OutlineInputBorder(borderSide: BorderSide(color: Colors.black, width:2))),
    );
  }

  static ThemeData fromMode(String mode, {bool darkSystem=false}) {
    switch(mode) {
      case 'light': return light();
      case 'dark': return dark();
      case 'eink': return eink();
      case 'system':
      default:
        return darkSystem ? dark() : light();
    }
  }
}
