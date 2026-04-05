import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/config_provider.dart';
import '../providers/recipe_provider.dart';
import '../widgets/recipe_card.dart';

class RecipeListScreen extends StatefulWidget {
  const RecipeListScreen({super.key});

  @override
  State<RecipeListScreen> createState() => _RecipeListScreenState();
}

class _RecipeListScreenState extends State<RecipeListScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _checkAndLoad();
    });
  }

  void _checkAndLoad() {
    if (!mounted) return;
    final provider = context.read<RecipeProvider?>();
    if (provider != null && provider.recipes.isEmpty && !provider.isLoading) {
      provider.loadRecipes();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Sandwitches Recipes'),
        elevation: 0,
        actions: [
          IconButton(
            icon: const Icon(Icons.settings),
            onPressed: () {
              showDialog(
                context: context,
                builder: (context) => AlertDialog(
                  title: const Text('Settings'),
                  content: const Text(
                      'Do you want to disconnect from this instance?'),
                  actions: [
                    TextButton(
                      onPressed: () => Navigator.pop(context),
                      child: const Text('Cancel'),
                    ),
                    TextButton(
                      onPressed: () {
                        context.read<ConfigProvider>().clear();
                        Navigator.pop(context);
                      },
                      child: const Text('Disconnect',
                          style: TextStyle(color: Colors.red)),
                    ),
                  ],
                ),
              );
            },
          ),
        ],
      ),
      body: Consumer<RecipeProvider?>(
        builder: (context, provider, child) {
          if (provider == null) {
            return const Center(child: CircularProgressIndicator());
          }

          // If provider became available but hasn't loaded yet
          if (provider.recipes.isEmpty &&
              !provider.isLoading &&
              provider.errorMessage == null) {
            WidgetsBinding.instance
                .addPostFrameCallback((_) => _checkAndLoad());
            return const Center(child: CircularProgressIndicator());
          }

          if (provider.isLoading && provider.recipes.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }

          if (provider.errorMessage != null) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text('Error: ${provider.errorMessage}'),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () => provider.loadRecipes(),
                    child: const Text('Retry'),
                  ),
                ],
              ),
            );
          }

          if (provider.recipes.isEmpty) {
            return const Center(child: Text('No recipes found.'));
          }

          return RefreshIndicator(
            onRefresh: () => provider.loadRecipes(),
            child: ListView.builder(
              itemCount: provider.recipes.length,
              itemBuilder: (context, index) {
                return RecipeCard(recipe: provider.recipes[index]);
              },
            ),
          );
        },
      ),
    );
  }
}
