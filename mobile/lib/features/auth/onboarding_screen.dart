import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:dio/dio.dart';
import '../../core/providers.dart';

class OnboardingScreen extends ConsumerStatefulWidget {
  const OnboardingScreen({super.key});
  @override
  ConsumerState<OnboardingScreen> createState()=> _OnboardingState();
}
class _OnboardingState extends ConsumerState<OnboardingScreen> {
  final ctrl = TextEditingController();
  bool loading=false;
  String? error;
  @override
  void initState(){
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) async {
      final s = ref.read(storageProvider);
      final saved = await s.getBaseUrl();
      if (saved!=null) {
        ctrl.text = saved;
        ref.read(baseUrlProvider.notifier).state = saved;
      }
    });
  }
  Future<void> validateAndSave() async {
    final url = ctrl.text.trim();
    if (url.isEmpty) { setState(()=> error='Please enter server URL'); return; }
    Uri? uri;
    try { uri = Uri.parse(url); if (!uri.hasScheme) throw FormatException(); } catch(_){ setState(()=> error='Invalid URL'); return; }
    setState(()=> loading=true);
    final base = url.endsWith('/') ? url.substring(0,url.length-1) : url;
    try {
      final dio = Dio(BaseOptions(connectTimeout: const Duration(seconds:5), receiveTimeout: const Duration(seconds:5), validateStatus: (s)=> s!=null && s<500));
      final r = await dio.get('$base/api/v1/ping');
      if (r.statusCode!=null && r.statusCode! <400) {
        await ref.read(storageProvider).setBaseUrl(base);
        ref.read(baseUrlProvider.notifier).state = base;
        if (mounted) context.go('/login');
      } else {
        setState(()=> error='Server did not respond with OK (${r.statusCode})');
      }
    } catch(e){
      setState(()=> error='Could not reach server: $e');
    } finally { if(mounted) setState(()=> loading=false); }
  }
  @override
  Widget build(BuildContext context){
    return Scaffold(
      appBar: AppBar(title: const Text('Welcome to Sandwitches')),
      body: Padding(padding: const EdgeInsets.all(16), child: Column(crossAxisAlignment: CrossAxisAlignment.stretch, children:[
        const Text('Enter your self-hosted server URL'),
        const SizedBox(height:12),
        TextField(controller: ctrl, decoration: InputDecoration(labelText: 'Server URL', hintText: 'https://example.com', errorText: error, border: const OutlineInputBorder()), keyboardType: TextInputType.url, onSubmitted: (_)=> validateAndSave()),
        const SizedBox(height:12),
        ElevatedButton(onPressed: loading? null: validateAndSave, child: loading? const SizedBox(height:20,width:20,child:CircularProgressIndicator(strokeWidth:2)) : const Text('Continue')),
      ])),
    );
  }
}
