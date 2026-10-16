import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sandwitches/features/auth/onboarding_screen.dart';

void main(){
  testWidgets('onboarding validates empty and invalid url', (tester) async {
    await tester.pumpWidget(const ProviderScope(child: MaterialApp(home: OnboardingScreen())));
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
