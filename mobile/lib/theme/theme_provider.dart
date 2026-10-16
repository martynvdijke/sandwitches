import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../core/providers.dart';
import '../core/storage.dart';

final themeNotifierProvider = StateNotifierProvider<ThemeNotifier, String>((ref){
  return ThemeNotifier(ref.watch(storageProvider));
});

class ThemeNotifier extends StateNotifier<String> {
  final AppStorage _storage;
  ThemeNotifier(this._storage): super('system') { _load(); }
  Future<void> _load() async {
    final v = await _storage.getTheme();
    if (v!=null) state=v;
  }
  Future<void> setTheme(String m) async {
    state=m;
    await _storage.setTheme(m);
  }
}
