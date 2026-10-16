import 'package:go_router/go_router.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../features/auth/onboarding_screen.dart';
import '../features/auth/login_screen.dart';
import '../features/home/home_screen.dart';
import '../features/recipe/recipe_detail_screen.dart';
import '../features/recipe/cooking_screen.dart';
import '../features/cart/cart_screen.dart';
import '../features/orders/order_tracker_screen.dart';
import '../features/orders/orders_screen.dart';
import '../features/admin/admin_queue_screen.dart';
import '../features/shell.dart';

final appRouterProvider = Provider<GoRouter>((ref){
  return GoRouter(
    initialLocation: '/onboarding',
    routes: [
      GoRoute(path: '/onboarding', builder: (c,s)=> const OnboardingScreen()),
      GoRoute(path: '/login', builder: (c,s)=> const LoginScreen()),
      ShellRoute(builder: (c,s,child)=> AppShell(child: child), routes: [
        GoRoute(path: '/', builder: (c,s)=> const HomeScreen()),
        GoRoute(path: '/cart', builder: (c,s)=> const CartScreen()),
        GoRoute(path: '/orders', builder: (c,s)=> const OrdersScreen()),
        GoRoute(path: '/admin', builder: (c,s)=> const AdminQueueScreen()),
      ]),
      GoRoute(path: '/recipe/:id', builder: (c,s){
        final id = int.parse(s.pathParameters['id']!);
        return RecipeDetailScreen(recipeId: id);
      }),
      GoRoute(path: '/recipe/:id/cooking', builder: (c,s){
        final id = int.parse(s.pathParameters['id']!);
        return CookingScreen(recipeId: id);
      }),
      GoRoute(path: '/orders/track/:token', builder: (c,s){
        final token = s.pathParameters['token']!;
        return OrderTrackerScreen(token: token);
      }),
    ],
  );
});
