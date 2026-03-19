import 'package:flutter/material.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../models/recipe.dart';

class RecipeDetailScreen extends StatelessWidget {
  final Recipe recipe;

  const RecipeDetailScreen({super.key, required this.recipe});

  @override
  Widget build(BuildContext context) {
    final String imageUrl = recipe.image != null
        ? (recipe.image!.startsWith('http')
            ? recipe.image!
            : 'http://localhost:8000${recipe.image}')
        : '';

    return Scaffold(
      body: CustomScrollView(
        slivers: [
          SliverAppBar(
            expandedHeight: 300,
            pinned: true,
            flexibleSpace: FlexibleSpaceBar(
              title: Text(
                recipe.title,
                style: const TextStyle(
                  color: Colors.white,
                  shadows: [Shadow(color: Colors.black, blurRadius: 4)],
                ),
              ),
              background: imageUrl.isNotEmpty
                  ? CachedNetworkImage(
                      imageUrl: imageUrl,
                      fit: BoxFit.cover,
                      errorWidget: (context, url, error) => Container(
                        color: Colors.grey[300],
                        child: const Icon(Icons.broken_image, size: 100),
                      ),
                    )
                  : Container(color: Colors.grey[400]),
            ),
          ),
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (recipe.description != null) ...[
                    Text(
                      recipe.description!,
                      style: Theme.of(context).textTheme.bodyLarge,
                    ),
                    const SizedBox(height: 24),
                  ],
                  _buildSectionTitle(context, 'Ingredients'),
                  const SizedBox(height: 8),
                  Text(recipe.ingredients ?? 'No ingredients listed.'),
                  const SizedBox(height: 24),
                  _buildSectionTitle(context, 'Instructions'),
                  const SizedBox(height: 8),
                  Text(recipe.instructions ?? 'No instructions listed.'),
                  const SizedBox(height: 32),
                  if (recipe.tags.isNotEmpty)
                    Wrap(
                      spacing: 8,
                      children: recipe.tags
                          .map((tag) => Chip(
                                label: Text(tag.name),
                                backgroundColor: Colors.orange[50],
                              ))
                          .toList(),
                    ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSectionTitle(BuildContext context, String title) {
    return Text(
      title,
      style: Theme.of(context).textTheme.headlineSmall?.copyWith(
            fontWeight: FontWeight.bold,
            color: Colors.orange[800],
          ),
    );
  }
}
