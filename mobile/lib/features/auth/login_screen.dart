import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:dio/dio.dart';
import 'package:dio_cookie_manager/dio_cookie_manager.dart';
import 'package:flutter_web_auth_2/flutter_web_auth_2.dart';
import '../../core/providers.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});
  @override
  ConsumerState<LoginScreen> createState()=> _LoginState();
}
class _LoginState extends ConsumerState<LoginScreen> {
  final userCtrl = TextEditingController();
  final passCtrl = TextEditingController();
  bool loading=false;
  String? error;
  bool oidcLoading=false;

  Future<void> doLogin() async {
    final base = ref.read(baseUrlProvider) ?? await ref.read(storageProvider).getBaseUrl() ?? '';
    if (base.isEmpty) { setState(()=> error='No server URL configured'); return; }
    setState(()=> loading=true);
    try {
      final dio = Dio(BaseOptions(
        baseUrl: base.endsWith('/') ? base.substring(0,base.length-1) : base,
        followRedirects: false,
        validateStatus: (s)=> s!=null && s<500,
        connectTimeout: const Duration(seconds:10),
        receiveTimeout: const Duration(seconds:10),
      ));
      final jar = ref.read(persistCookieJarProvider);
      dio.interceptors.add(CookieManager(jar));
      final r = await dio.post('$base/login',
        data: {'username': userCtrl.text.trim(), 'password': passCtrl.text},
        options: Options(contentType: Headers.formUrlEncodedContentType, followRedirects: false, validateStatus: (s)=> s!=null && s<500),
      );
      // success if 302 redirect or 200 with cookie?
      if (r.statusCode!=null && r.statusCode! >=300 && r.statusCode! <400) {
        if (mounted) context.go('/');
      } else if (r.statusCode==200) {
        // check if login failed: server re-renders login HTML -> likely no session
        // Try fetching /api/v1/me
        try {
          final me = await dio.get('$base/api/v1/me');
          if (me.statusCode==200) {
            if (mounted) context.go('/');
          } else {
            setState(()=> error='Invalid credentials');
          }
        } catch(_){
          setState(()=> error='Invalid credentials');
        }
      } else {
        setState(()=> error='Login failed (${r.statusCode})');
      }
    } catch(e){
      setState(()=> error='Login error: $e');
    } finally { if(mounted) setState(()=> loading=false); }
  }

  Future<void> doOidc() async {
    final baseRaw = ref.read(baseUrlProvider) ?? await ref.read(storageProvider).getBaseUrl() ?? '';
    final base = baseRaw.endsWith('/') ? baseRaw.substring(0,baseRaw.length-1) : baseRaw;
    if (base.isEmpty) { setState(()=> error='No server URL'); return; }
    setState(()=> oidcLoading=true);
    try {
      const callbackScheme = 'sandwitches';
      final result = await FlutterWebAuth2.authenticate(
        url: '$base/login/oidc?mobile=1',
        callbackUrlScheme: callbackScheme,
      );
      final callback = Uri.parse(result);
      if (callback.scheme != callbackScheme || callback.host != 'auth') {
        setState(()=> error='OIDC login returned an unexpected response');
        return;
      }
      final oidcError = callback.queryParameters['error'];
      if (oidcError != null && oidcError.isNotEmpty) {
        setState(()=> error=_oidcErrorMessage(oidcError));
        return;
      }
      final code = callback.queryParameters['code'];
      if (code == null || code.isEmpty) {
        setState(()=> error='OIDC login did not return a code');
        return;
      }
      final dio = ref.read(dioProvider);
      final exchange = await dio.post('$base/api/auth/mobile/exchange',
        data: {'code': code},
        options: Options(contentType: Headers.jsonContentType, followRedirects: false),
      );
      if (exchange.statusCode != 200) {
        setState(()=> error='OIDC login could not be completed');
        return;
      }
      final me = await dio.get('$base/api/v1/me');
      if (me.statusCode==200) {
        if (mounted) context.go('/');
      } else {
        setState(()=> error='OIDC login failed');
      }
    } on PlatformException catch(e) {
      if (e.code != 'CANCELED') {
        setState(()=> error='OIDC login failed. Please try again.');
      }
    } catch(_) {
      setState(()=> error='OIDC login failed. Please try again.');
    } finally { if(mounted) setState(()=> oidcLoading=false); }
  }

  String _oidcErrorMessage(String code) {
    switch (code) {
      case 'oidc_expired':
        return 'OIDC login expired. Please try again.';
      case 'oidc_state':
      case 'oidc_nonce':
        return 'OIDC login could not be verified. Please try again.';
      case 'oidc_email':
        return 'Your OIDC account needs a verified email address.';
      case 'oidc_handoff':
        return 'OIDC login could not be handed to the app. Please try again.';
      default:
        return 'OIDC login failed ($code)';
    }
  }

  Future<void> logout() async {
    await ref.read(storageProvider).clearAuth();
    await ref.read(persistCookieJarProvider).deleteAll();
    if (mounted) context.go('/login');
  }

  @override
  Widget build(BuildContext context){
    final base = ref.watch(baseUrlProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('Login')),
      body: Padding(padding: const EdgeInsets.all(16), child: ListView(children:[
        if (base!=null) Text('Server: $base', style: Theme.of(context).textTheme.bodySmall),
        TextButton(onPressed: ()=> context.go('/onboarding'), child: const Text('Change server URL')),
        const SizedBox(height:12),
        TextField(controller: userCtrl, decoration: const InputDecoration(labelText: 'Username', border: OutlineInputBorder())),
        const SizedBox(height:12),
        TextField(controller: passCtrl, obscureText: true, decoration: const InputDecoration(labelText: 'Password', border: OutlineInputBorder())),
        if (error!=null) Padding(padding: const EdgeInsets.only(top:8), child: Text(error!, style: const TextStyle(color: Colors.red))),
        const SizedBox(height:12),
        ElevatedButton(onPressed: loading? null: doLogin, child: loading? const SizedBox(height:20,width:20,child:CircularProgressIndicator(strokeWidth:2)) : const Text('Login')),
        const SizedBox(height:12),
        OutlinedButton(onPressed: oidcLoading? null: doOidc, child: oidcLoading? const SizedBox(height:20,width:20,child:CircularProgressIndicator(strokeWidth:2)) : const Text('Login with OIDC')),
        const SizedBox(height:12),
        TextButton(onPressed: logout, child: const Text('Clear session (logout)')),
      ])),
    );
  }
}
