import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:sandwitches/main.dart' as app;

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('app boots and renders its first screen', (tester) async {
    await Future<void>(() => app.main());

    final root = find.byWidgetPredicate((widget) => widget is WidgetsApp);
    await _pumpUntil(tester, root, const Duration(seconds: 30));

    expect(root, findsWidgets,
        reason: 'The app must mount a WidgetsApp after launch');
    expect(tester.takeException(), isNull);
  });
}

/// Pumps frames until [finder] matches, unlike pumpAndSettle this also works on
/// screens with infinite animations (splash spinners, pulsing headers).
Future<void> _pumpUntil(
  WidgetTester tester,
  Finder finder,
  Duration timeout,
) async {
  final deadline = DateTime.now().add(timeout);
  while (DateTime.now().isBefore(deadline)) {
    if (tester.any(finder)) return;
    await tester.pump(const Duration(milliseconds: 100));
  }
}
