class User {
  final String username;
  final String? firstName;
  final String? lastName;
  final String? avatar;

  User({
    required this.username,
    this.firstName,
    this.lastName,
    this.avatar,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      username: json['username'],
      firstName: json['first_name'],
      lastName: json['last_name'],
      avatar: json['avatar'],
    );
  }

  String get displayName =>
      (firstName != null && firstName!.isNotEmpty) ? firstName! : username;
}
