import 'package:flutter_test/flutter_test.dart';
import 'package:sandwitches/core/providers.dart';

void main() {
  group('DemoLoginConfig.fromJson', () {
    test('parses a valid enabled demo config', () {
      final config = DemoLoginConfig.fromJson({
        'demo': {
          'enabled': true,
          'username': 'demo-user',
          'password': 'demo-pass',
        },
      });

      expect(config.enabled, isTrue);
      expect(config.username, 'demo-user');
      expect(config.password, 'demo-pass');
    });

    test('parses a disabled demo config', () {
      final config = DemoLoginConfig.fromJson({
        'demo': {'enabled': false, 'username': 'demo', 'password': 'pass'},
      });

      expect(config.enabled, isFalse);
      expect(config.username, 'demo');
      expect(config.password, 'pass');
    });

    test('defaults to disabled when the demo object is missing', () {
      final config = DemoLoginConfig.fromJson({});

      expect(config.enabled, isFalse);
      expect(config.username, '');
      expect(config.password, '');
    });

    test('tolerates malformed values', () {
      final config = DemoLoginConfig.fromJson({
        'demo': {'enabled': 'yes', 'username': 42, 'password': null},
      });

      expect(config.enabled, isFalse);
      expect(config.username, '');
      expect(config.password, '');
    });

    test('tolerates a non-object demo value', () {
      final config = DemoLoginConfig.fromJson({'demo': 'enabled'});

      expect(config.enabled, isFalse);
      expect(config.username, '');
      expect(config.password, '');
    });
  });
}
