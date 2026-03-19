import 'tag.dart';
import 'user.dart';

class Recipe {
  final int id;
  final String title;
  final String? description;
  final String? ingredients;
  final String? instructions;
  final int? servings;
  final double? price;
  final String? image;
  final List<Tag> tags;
  final List<User> favoritedBy;

  Recipe({
    required this.id,
    required this.title,
    this.description,
    this.ingredients,
    this.instructions,
    this.servings,
    this.price,
    this.image,
    required this.tags,
    required this.favoritedBy,
  });

  factory Recipe.fromJson(Map<String, dynamic> json) {
    return Recipe(
      id: json['id'],
      title: json['title'],
      description: json['description'],
      ingredients: json['ingredients'],
      instructions: json['instructions'],
      servings: json['servings'],
      price: json['price']?.toDouble(),
      image: json['image'],
      tags: (json['tags'] as List?)?.map((t) => Tag.fromJson(t)).toList() ?? [],
      favoritedBy: (json['favorited_by'] as List?)
              ?.map((u) => User.fromJson(u))
              .toList() ??
          [],
    );
  }
}
