import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

class AppShell extends StatelessWidget {
  final Widget child;
  const AppShell({super.key, required this.child});
  @override
  Widget build(BuildContext context) {
    final loc = GoRouterState.of(context).uri.toString();
    int idx = 0;
    if (loc.startsWith('/cart')) {
      idx=1;
    } else if (loc.startsWith('/orders')) {
      idx=2;
    } else if (loc.startsWith('/admin')) {
      idx=3;
    }
    return Scaffold(
      body: child,
      bottomNavigationBar: NavigationBar(
        selectedIndex: idx,
        onDestinationSelected: (i){
          switch(i){
            case 0: context.go('/'); break;
            case 1: context.go('/cart'); break;
            case 2: context.go('/orders'); break;
            case 3: context.go('/admin'); break;
          }
        },
        destinations: const [
          NavigationDestination(icon: Icon(Icons.home), label: 'Home'),
          NavigationDestination(icon: Icon(Icons.shopping_cart), label: 'Cart'),
          NavigationDestination(icon: Icon(Icons.receipt_long), label: 'Orders'),
          NavigationDestination(icon: Icon(Icons.admin_panel_settings), label: 'Admin'),
        ],
      ),
    );
  }
}
