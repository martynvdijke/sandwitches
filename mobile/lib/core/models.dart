class Tag {
  final int id;
  final String name;
  final String slug;
  Tag({required this.id, required this.name, required this.slug});
  factory Tag.fromJson(Map<String,dynamic> j) => Tag(id: j['id'] as int, name: j['name'] as String, slug: j['slug'] as String);
}

class Recipe {
  final int id; final String title; final String slug; final String description;
  final String ingredients; final String instructions; final int servings;
  final String price; final String image; final String imageThumbnail;
  final String imageSmall; final String imageMedium; final String imageLarge;
  final bool isHighlighted; final bool isApproved;
  final int? prepTime; final int? cookTime; final int? calories;
  final int? uploadedBy; final List<Tag> tags; final List<int> favoritedBy;
  final double? averageRating; final int? dailyOrdersCount;
  final String? createdAt; final String? updatedAt;
  Recipe({
    required this.id, required this.title, required this.slug,
    required this.description, required this.ingredients, required this.instructions,
    required this.servings, required this.price, required this.image,
    this.imageThumbnail='', this.imageSmall='', this.imageMedium='', this.imageLarge='',
    this.isHighlighted=false, this.isApproved=true,
    this.prepTime, this.cookTime, this.calories, this.uploadedBy,
    this.tags=const [], this.favoritedBy=const [], this.averageRating, this.dailyOrdersCount,
    this.createdAt, this.updatedAt,
  });
  factory Recipe.fromJson(Map<String,dynamic> j) {
    List<Tag> tags = [];
    if (j['tags'] is List) tags = (j['tags'] as List).map((e)=>Tag.fromJson(e as Map<String,dynamic>)).toList();
    List<int> fav = [];
    if (j['favorited_by'] is List) fav = (j['favorited_by'] as List).map((e)=> (e as num).toInt()).toList();
    double? avg;
    if (j['average_rating'] != null) avg = (j['average_rating'] as num).toDouble();
    return Recipe(
      id: (j['id'] as num).toInt(), title: j['title'] as String? ?? '', slug: j['slug'] as String? ?? '',
      description: j['description'] as String? ?? '', ingredients: j['ingredients'] as String? ?? '',
      instructions: j['instructions'] as String? ?? '', servings: (j['servings'] as num?)?.toInt() ?? 1,
      price: j['price']?.toString() ?? '0', image: j['image'] as String? ?? '',
      imageThumbnail: j['image_thumbnail'] as String? ?? '', imageSmall: j['image_small'] as String? ?? '',
      imageMedium: j['image_medium'] as String? ?? '', imageLarge: j['image_large'] as String? ?? '',
      isHighlighted: j['is_highlighted'] as bool? ?? false, isApproved: j['is_approved'] as bool? ?? true,
      prepTime: (j['prep_time'] as num?)?.toInt(), cookTime: (j['cook_time'] as num?)?.toInt(),
      calories: (j['calories'] as num?)?.toInt(), uploadedBy: (j['uploaded_by'] as num?)?.toInt(),
      tags: tags, favoritedBy: fav, averageRating: avg,
      dailyOrdersCount: (j['daily_orders_count'] as num?)?.toInt(),
      createdAt: j['created_at'] as String?, updatedAt: j['updated_at'] as String?,
    );
  }
}

class Rating {
  final int id; final int recipe; final int score; final String comment;
  final String? image; final Map<String,dynamic>? user;
  final String? createdAt; final String? updatedAt;
  Rating({required this.id, required this.recipe, required this.score, required this.comment, this.image, this.user, this.createdAt, this.updatedAt});
  factory Rating.fromJson(Map<String,dynamic> j) => Rating(
    id: (j['id'] as num).toInt(), recipe: (j['recipe'] as num?)?.toInt() ?? 0,
    score: (j['score'] as num).toInt(), comment: j['comment'] as String? ?? '',
    image: j['image'] as String?, user: j['user'] as Map<String,dynamic>?,
    createdAt: j['created_at'] as String?, updatedAt: j['updated_at'] as String?,
  );
}

class ScaledIngredient {
  final String originalLine; final String scaledLine; final double? quantity; final String unit; final String name;
  ScaledIngredient({required this.originalLine, required this.scaledLine, this.quantity, required this.unit, required this.name});
  factory ScaledIngredient.fromJson(Map<String,dynamic> j) => ScaledIngredient(
    originalLine: j['original_line'] as String? ?? '', scaledLine: j['scaled_line'] as String? ?? '',
    quantity: (j['quantity'] as num?)?.toDouble(), unit: j['unit'] as String? ?? '', name: j['name'] as String? ?? '',
  );
}

class CartItem {
  final int id; final Recipe recipe; final int quantity;
  CartItem({required this.id, required this.recipe, required this.quantity});
  factory CartItem.fromJson(Map<String,dynamic> j) => CartItem(
    id: (j['id'] as num).toInt(), recipe: Recipe.fromJson(j['recipe'] as Map<String,dynamic>), quantity: (j['quantity'] as num).toInt(),
  );
}

class Order {
  final int id; final String status; final String totalPrice; final String createdAt; final String updatedAt; final String trackingToken;
  Order({required this.id, required this.status, required this.totalPrice, required this.createdAt, required this.updatedAt, required this.trackingToken});
  factory Order.fromJson(Map<String,dynamic> j) => Order(
    id: (j['id'] as num).toInt(), status: j['status'] as String? ?? 'PENDING',
    totalPrice: j['total_price']?.toString() ?? '0', createdAt: j['created_at'] as String? ?? '', updatedAt: j['updated_at'] as String? ?? '',
    trackingToken: j['tracking_token'] as String? ?? j['trackingToken'] as String? ?? '',
  );
}

class OrderDetail {
  final int id; final String status; final String totalPrice; final String createdAt; final String updatedAt; final String trackingToken;
  final List<OrderItem> items;
  OrderDetail({required this.id, required this.status, required this.totalPrice, required this.createdAt, required this.updatedAt, required this.trackingToken, required this.items});
  factory OrderDetail.fromJson(Map<String,dynamic> j) {
    List<OrderItem> items = [];
    if (j['items'] is List) items = (j['items'] as List).map((e)=>OrderItem.fromJson(e as Map<String,dynamic>)).toList();
    return OrderDetail(
      id: (j['id'] as num).toInt(), status: j['status'] as String? ?? 'PENDING',
      totalPrice: j['total_price']?.toString() ?? '0', createdAt: j['created_at'] as String? ?? '', updatedAt: j['updated_at'] as String? ?? '',
      trackingToken: j['tracking_token'] as String? ?? '', items: items,
    );
  }
}

class OrderItem {
  final int id; final int quantity; final String price; final Recipe recipe;
  OrderItem({required this.id, required this.quantity, required this.price, required this.recipe});
  factory OrderItem.fromJson(Map<String,dynamic> j) => OrderItem(
    id: (j['id'] as num).toInt(), quantity: (j['quantity'] as num).toInt(), price: j['price']?.toString() ?? '0',
    recipe: Recipe.fromJson(j['recipe'] as Map<String,dynamic>),
  );
}

String buildThumbUrl(String baseUrl, String imagePath, {int w = 800}) {
  if (imagePath.isEmpty) {
    return '';
  }
  if (imagePath.startsWith('http://') || imagePath.startsWith('https://') || imagePath.startsWith('data:')) {
    return imagePath;
  }
  String path = imagePath;
  if (path.startsWith('/media/')) {
    path = path.substring('/media/'.length);
  }
  else if (path.startsWith('media/')) {
    path = path.substring('media/'.length);
  }
  else if (path.startsWith('/')) {
    path = path.substring(1);
  }
  final base = baseUrl.endsWith('/') ? baseUrl.substring(0, baseUrl.length-1) : baseUrl;
  return '$base/thumb/$path?w=$w';
}
