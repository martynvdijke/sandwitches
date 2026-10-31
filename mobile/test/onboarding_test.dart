import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sandwitches/core/config.dart';
import 'package:sandwitches/core/providers.dart';
import 'package:sandwitches/core/storage.dart';
import 'package:sandwitches/features/auth/onboarding_screen.dart';

class _FakeStorage extends AppStorage {
  _FakeStorage(this.saved);
  final String? saved;
  String? written;

  @override
  Future<String?> getBaseUrl() async => saved;

  @override
  Future<void> setBaseUrl(String v) async {
    written = v;
  }
}

Future<void> _pumpOnboarding(WidgetTester tester, AppStorage storage) async {
  await tester.pumpWidget(
    ProviderScope(
      overrides: [storageProvider.overrideWithValue(storage)],
      child: const MaterialApp(home: OnboardingScreen()),
    ),
  );
  await tester.pumpAndSettle();
}

String? _fieldText(WidgetTester tester) =>
    tester.widget<TextField>(find.byType(TextField)).controller?.text;

void main() {
  testWidgets('onboarding prefills the default server when none is saved', (
    tester,
  ) async {
    final storage = _FakeStorage(null);
    await _pumpOnboarding(tester, storage);

    expect(_fieldText(tester), kDefaultBaseUrl);
  });

  testWidgets('onboarding keeps a previously saved server URL', (tester) async {
    final storage = _FakeStorage('https://recipes.example.com');
    await _pumpOnboarding(tester, storage);

    expect(_fieldText(tester), 'https://recipes.example.com');
  });

  testWidgets('onboarding validates empty and invalid url', (tester) async {
    await _pumpOnboarding(tester, _FakeStorage(null));

    await tester.enterText(find.byType(TextField), '');
    await tester.tap(find.text('Continue'));
    await tester.pump();
    expect(find.text('Please enter server URL'), findsOneWidget);

    await tester.enterText(find.byType(TextField), 'not a url');
    await tester.tap(find.text('Continue'));
    await tester.pump();
    // Should show invalid or attempt network; at least error set
    expect(find.textContaining('Invalid'), findsOneWidget);
  });
}
